# Xây dựng Module Chạy và Dừng Script
Dựa vào yêu cầu của bạn, tôi sẽ xây dựng module để thực hiện hai chức năng chính:

1. API run file script sử dụng SSE để stream dữ liệu console log về client
2. API stop tiến trình chạy file script


## Các bước thực hiện 
1. Tạo Model cho Process
2. Tạo Repository cho Process
3. Tạo Service cho Process
4. Tạo Controller cho Process
5. Tạo Router cho Process

## Lưu ý về triển khai
1. Bảo mật :
   
   - Chúng ta đang chạy script trong môi trường của server, điều này tiềm ẩn rủi ro bảo mật
   - Trong môi trường production, nên sử dụng container hoặc sandbox để cách ly script
   - Nên giới hạn tài nguyên và thời gian chạy script
2. Xử lý SSE :
   
   - Đảm bảo đóng kết nối khi script kết thúc hoặc bị dừng
   - Xử lý trường hợp client disconnect
3. Quản lý tiến trình :
   
   - Lưu trữ thông tin tiến trình trong database
   - Cơ chế để kill tiến trình khi nhận lệnh stop
   - Xử lý trường hợp tiến trình "zombie"
4. Cải tiến trong tương lai :
   
   - Thêm giới hạn thời gian chạy script
   - Thêm giới hạn tài nguyên (CPU, memory)
   - Thêm tính năng lưu lịch sử chạy script
   - Thêm tính năng schedule chạy script
## Tóm tắt API
1. Chạy Script :
   
   - POST /api/scripts/:id/run - Chạy script và stream output qua SSE
   - Request body: { "args": ["arg1", "arg2", ...] } (tùy chọn)
   - Response: SSE stream với output của script
2. Dừng Process :
   
   - POST /api/processes/:id/stop - Dừng một process đang chạy
   - Response: { "message": "Process stopped successfully" }
3. Lấy danh sách Process :
   
   - GET /api/processes - Lấy danh sách các process của user
   - Response: Danh sách các process
Với module này, bạn đã hoàn thành các yêu cầu API của giai đoạn 1 trong dự án.