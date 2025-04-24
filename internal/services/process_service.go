package services

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"scripts-management/internal/models"
	"scripts-management/internal/repository"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

type ProcessService struct {
	processRepo   *repository.ProcessRepository
	scriptRepo    *repository.ScriptRepository
	scriptService *ScriptService
	logger        *zap.Logger
	processes     map[primitive.ObjectID]*models.Process
	mu            sync.Mutex
}

func NewProcessService(
	processRepo *repository.ProcessRepository,
	scriptRepo *repository.ScriptRepository,
	scriptService *ScriptService,
	logger *zap.Logger,
) *ProcessService {
	return &ProcessService{
		processRepo:   processRepo,
		scriptRepo:    scriptRepo,
		scriptService: scriptService,
		logger:        logger,
		processes:     make(map[primitive.ObjectID]*models.Process),
		mu:            sync.Mutex{},
	}
}

func (s *ProcessService) RunScript(ctx context.Context, userID primitive.ObjectID, scriptID primitive.ObjectID, Args []string) (*models.Process, error) {
	// Kiểm tra quyền truy cập script
	script, err := s.scriptService.GetScriptByID(ctx, userID, scriptID)
	if err != nil {
		return nil, fmt.Errorf("không thể truy cập script: %w", err)
	}

	// Kiểm tra xem script có đang chạy không
	runningProcess, err := s.processRepo.FindRunningByScriptID(ctx, scriptID)
	if err != nil {
		return nil, fmt.Errorf("lỗi khi kiểm tra tiến trình đang chạy: %w", err)
	}

	if runningProcess != nil {
		return nil, errors.New("script đang chạy, vui lòng dừng tiến trình hiện tại trước khi chạy lại")
	}

	// Tạo thư mục tạm thời để lưu script
	tempDir, err := os.MkdirTemp("", "script-*")
	if err != nil {
		return nil, fmt.Errorf("không thể tạo thư mục tạm thời: %w", err)
	}

	var cmd *exec.Cmd
	var scriptPath string

	// Tạo file script dựa vào loại
	switch script.Type {
	case "python":
		scriptPath = filepath.Join(tempDir, "script.py")
		if err := os.WriteFile(scriptPath, []byte(script.Content), 0644); err != nil {
			os.RemoveAll(tempDir)
			return nil, fmt.Errorf("không thể tạo file script: %w", err)
		}
		cmd = exec.Command("python3", scriptPath)

	case "golang":
		scriptPath = filepath.Join(tempDir, "script.go")
		if err := os.WriteFile(scriptPath, []byte(script.Content), 0644); err != nil {
			os.RemoveAll(tempDir)
			return nil, fmt.Errorf("không thể tạo file script: %w", err)
		}
		cmd = exec.Command("go", "run", scriptPath)

	default:
		os.RemoveAll(tempDir)
		return nil, fmt.Errorf("loại script không được hỗ trợ: %s", script.Type)
	}

	// Tạo process mới
	process := &models.Process{
		ID:        primitive.NewObjectID(),
		ScriptID:  scriptID,
		UserID:    userID,
		Status:    models.ProcessStatusRunning,
		StartTime: time.Now(),
		Cmd:       cmd,
	}

	// Lưu process vào DB
	if err := s.processRepo.Create(ctx, process); err != nil {
		os.RemoveAll(tempDir)
		return nil, fmt.Errorf("không thể lưu thông tin tiến trình: %w", err)
	}

	// Lưu process vào memory
	s.mu.Lock()
	s.processes[process.ID] = process
	s.mu.Unlock()

	return process, nil
}

func (s *ProcessService) StreamProcessOutput(c *fiber.Ctx, processID primitive.ObjectID) error {
	// Thiết lập SSE
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	// Lấy process từ memory
	s.mu.Lock()
	process, exists := s.processes[processID]
	s.mu.Unlock()

	if !exists {
		// Kiểm tra trong DB
		ctx := c.Context()
		var err error
		process, err = s.processRepo.FindByID(ctx, processID)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Không tìm thấy tiến trình",
			})
		}

		if process.Status != models.ProcessStatusRunning {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Tiến trình không đang chạy",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Không thể stream output của tiến trình này",
		})
	}

	// Tạo pipe cho stdout và stderr
	stdoutPipe, err := process.Cmd.StdoutPipe()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Không thể tạo pipe cho stdout: %v", err),
		})
	}

	stderrPipe, err := process.Cmd.StderrPipe()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Không thể tạo pipe cho stderr: %v", err),
		})
	}

	// Bắt đầu chạy command
	if err := process.Cmd.Start(); err != nil {
		s.mu.Lock()
		delete(s.processes, processID)
		s.mu.Unlock()

		// Cập nhật trạng thái trong DB
		ctx := c.Context()
		errMsg := err.Error()
		s.processRepo.UpdateStatus(ctx, processID, models.ProcessStatusError, nil, errMsg)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Không thể chạy script: %v", err),
		})
	}

	// Lưu PID
	process.PID = process.Cmd.Process.Pid
	ctx := c.Context()
	s.processRepo.Update(ctx, process)

	// Tạo channel để thông báo khi process kết thúc
	done := make(chan error, 1)
	go func() {
		done <- process.Cmd.Wait()
	}()

	// Tạo channel để đọc output
	outputChan := make(chan string)

	// Đọc stdout
	go func() {
		reader := bufio.NewReader(stdoutPipe)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					s.logger.Error("Lỗi khi đọc stdout", zap.Error(err))
				}
				break
			}
			outputChan <- fmt.Sprintf("data: %s\n\n", strings.TrimSpace(line))
		}
	}()

	// Đọc stderr
	go func() {
		reader := bufio.NewReader(stderrPipe)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					s.logger.Error("Lỗi khi đọc stderr", zap.Error(err))
				}
				break
			}
			outputChan <- fmt.Sprintf("data: [ERROR] %s\n\n", strings.TrimSpace(line))
		}
	}()

	// Thiết lập notifier khi client disconnect
	notifier := c.Context().Done()

	// Stream output
	// c.Context().SetBodyStreamWriter(func(w io.Writer) {
	// 	for {
	// 		select {
	// 		case <-notifier:
	// 			// Client đã ngắt kết nối
	// 			s.logger.Info("Client ngắt kết nối, dừng tiến trình", zap.String("processID", processID.Hex()))
	// 			s.StopProcess(c.Context(), processID)
	// 			return

	// 		case output := <-outputChan:
	// 			// Gửi output về client
	// 			fmt.Fprint(w, output)
	// 			if f, ok := w.(http.Flusher); ok {
	// 				f.Flush()
	// 			}

	// 		case err := <-done:
	// 			// Process đã kết thúc
	// 			s.mu.Lock()
	// 			delete(s.processes, processID)
	// 			s.mu.Unlock()

	// 			// Cập nhật trạng thái trong DB
	// 			exitCode := 0
	// 			errMsg := ""
	// 			if err != nil {
	// 				s.logger.Error("Process kết thúc với lỗi", zap.Error(err))
	// 				errMsg = err.Error()
	// 				if exitErr, ok := err.(*exec.ExitError); ok {
	// 					exitCode = exitErr.ExitCode()
	// 				}
	// 			}

	// 			status := models.ProcessStatusStopped
	// 			if errMsg != "" {
	// 				status = models.ProcessStatusError
	// 			}

	// 			s.processRepo.UpdateStatus(c.Context(), processID, status, &exitCode, errMsg)

	// 			// Gửi thông báo kết thúc
	// 			fmt.Fprintf(w, "event: end\ndata: {\"status\":\"%s\",\"exitCode\":%d,\"error\":\"%s\"}\n\n", status, exitCode, errMsg)
	// 			if f, ok := w.(http.Flusher); ok {
	// 				f.Flush()
	// 			}
	// 			return
	// 		}
	// 	}
	// })

	c.Context().Response().SetBodyStreamWriter(func(w *bufio.Writer) {
		for {
			select {
			case <-notifier:
				// Client đã ngắt kết nối
				s.logger.Info("Client ngắt kết nối, dừng tiến trình", zap.String("processID", processID.Hex()))
				s.StopProcess(c.Context(), processID)
				return

			case output := <-outputChan:
				// Gửi output về client
				fmt.Fprint(w, output)
				w.Flush()

			case err := <-done:
				// Process đã kết thúc
				s.mu.Lock()
				delete(s.processes, processID)
				s.mu.Unlock()

				// Cập nhật trạng thái trong DB
				exitCode := 0
				errMsg := ""
				if err != nil {
					s.logger.Error("Process kết thúc với lỗi", zap.Error(err))
					errMsg = err.Error()
					if exitErr, ok := err.(*exec.ExitError); ok {
						exitCode = exitErr.ExitCode()
					}
				}

				status := models.ProcessStatusStopped
				if errMsg != "" {
					status = models.ProcessStatusError
				}

				s.processRepo.UpdateStatus(c.Context(), processID, status, &exitCode, errMsg)

				// Gửi thông báo kết thúc
				fmt.Fprintf(w, "event: end\ndata: {\"status\":\"%s\",\"exitCode\":%d,\"error\":\"%s\"}\n\n", status, exitCode, errMsg)
				w.Flush()
				return
			}
		}
	})

	return nil
}

func (s *ProcessService) StopProcess(ctx context.Context, processID primitive.ObjectID) error {
	// Kiểm tra process trong memory
	s.mu.Lock()
	process, exists := s.processes[processID]
	s.mu.Unlock()

	if !exists {
		// Kiểm tra trong DB
		var err error
		process, err = s.processRepo.FindByID(ctx, processID)
		if err != nil {
			return fmt.Errorf("không tìm thấy tiến trình: %w", err)
		}

		if process.Status != models.ProcessStatusRunning {
			return errors.New("tiến trình không đang chạy")
		}

		// Nếu process không có trong memory nhưng đang chạy trong DB
		// Thử kill process bằng PID
		if process.PID > 0 {
			proc, err := os.FindProcess(process.PID)
			if err == nil {
				// Gửi signal SIGTERM
				if err := proc.Signal(syscall.SIGTERM); err != nil {
					s.logger.Error("Không thể gửi SIGTERM đến process", zap.Int("pid", process.PID), zap.Error(err))

					// Thử SIGKILL nếu SIGTERM thất bại
					if err := proc.Kill(); err != nil {
						s.logger.Error("Không thể kill process", zap.Int("pid", process.PID), zap.Error(err))
					}
				}
			}
		}

		// Cập nhật trạng thái trong DB
		exitCode := -1
		s.processRepo.UpdateStatus(ctx, processID, models.ProcessStatusStopped, &exitCode, "Process bị dừng bởi người dùng")
		return nil
	}

	// Nếu process có trong memory
	if process.Cmd != nil && process.Cmd.Process != nil {
		// Gửi signal SIGTERM
		if err := process.Cmd.Process.Signal(syscall.SIGTERM); err != nil {
			s.logger.Error("Không thể gửi SIGTERM đến process", zap.Error(err))

			// Thử SIGKILL nếu SIGTERM thất bại
			if err := process.Cmd.Process.Kill(); err != nil {
				s.logger.Error("Không thể kill process", zap.Error(err))
				return fmt.Errorf("không thể dừng tiến trình: %w", err)
			}
		}
	}

	// Xóa process khỏi memory
	s.mu.Lock()
	delete(s.processes, processID)
	s.mu.Unlock()

	// Cập nhật trạng thái trong DB
	exitCode := -1
	s.processRepo.UpdateStatus(ctx, processID, models.ProcessStatusStopped, &exitCode, "Process bị dừng bởi người dùng")

	return nil
}

func (s *ProcessService) GetProcessByID(ctx context.Context, userID, processID primitive.ObjectID) (*models.Process, error) {
	process, err := s.processRepo.FindByID(ctx, processID)
	if err != nil {
		return nil, fmt.Errorf("không tìm thấy tiến trình: %w", err)
	}

	// Kiểm tra quyền truy cập
	if process.UserID != userID {
		// Kiểm tra xem user có quyền truy cập script không
		_, err := s.scriptService.GetScriptByID(ctx, userID, process.ScriptID)
		if err != nil {
			return nil, errors.New("không có quyền truy cập tiến trình này")
		}
	}

	return process, nil
}

func (s *ProcessService) GetProcessesByScriptID(ctx context.Context, userID, scriptID primitive.ObjectID) ([]*models.Process, error) {
	// Kiểm tra quyền truy cập script
	_, err := s.scriptService.GetScriptByID(ctx, userID, scriptID)
	if err != nil {
		return nil, fmt.Errorf("không thể truy cập script: %w", err)
	}

	processes, err := s.processRepo.FindByScriptID(ctx, scriptID)
	if err != nil {
		return nil, fmt.Errorf("lỗi khi lấy danh sách tiến trình: %w", err)
	}

	return processes, nil
}
