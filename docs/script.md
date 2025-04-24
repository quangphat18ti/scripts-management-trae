# Triển khai Module Quản lý Script và Chia sẻ Script
Tôi sẽ giúp bạn triển khai bộ module để quản lý script và chia sẻ script cho người khác. Dựa vào yêu cầu của bạn, chúng ta cần tạo các API CRUD để quản lý các file script Python/Golang, với các quy tắc phân quyền:

- User chỉ có quyền với file của chính mình
- Cho phép share file script với các user khác
- User được share không có quyền xóa file

## Tóm tắt API
Với các module đã triển khai, bạn sẽ có các API sau:

1. Quản lý Script :
   
   - POST /api/scripts - Tạo script mới
   - GET /api/scripts - Lấy danh sách script của user (bao gồm cả script được share)
   - GET /api/scripts/:id - Lấy chi tiết một script
   - PUT /api/scripts/:id - Cập nhật script (chỉ owner)
   - DELETE /api/scripts/:id - Xóa script (chỉ owner)
2. Chia sẻ Script :
   
   - POST /api/scripts/:id/share - Chia sẻ script với user khác
   - DELETE /api/scripts/:id/share/:userId - Hủy chia sẻ script
Các tính năng chính:

- Quản lý script Python/Golang
- Phân quyền: user chỉ có quyền với file của mình
- Chia sẻ script với user khác
- User được share không có quyền xóa file
Bạn có thể mở rộng thêm các tính năng như:

- Thêm tính năng chạy script
- Thêm lịch sử phiên bản script
- Thêm tính năng comment/feedback cho script được chia sẻ