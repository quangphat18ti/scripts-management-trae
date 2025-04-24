# Training Project

## Tổng quan toàn bộ quy trình 2 tháng

Technology:

- Ngôn ngữ: golang
- Framework: gofiber
- Database: mongodb

Tech stack:

- GoFiber
- Mongo
- Docker
- Zap
- DI pattern
- Module Base Structure

Nội dung:

- Thực hiện công cụ cho phép quản lý các script python / golang có trên server

### Thời gian

Giai đoạn 1: 1 tháng

- Khởi tạo base project
- Login bằng JWT sử dụng ecdsa
- Tạo tài khoản root với giá trị lấy từ ENV, cho phép tạo thêm các tài khoản từ tài khoản root. Có 2 loại tài khoản là admin và member: admin có quyền thêm, xóa, đổi mật khẩu member, admin ko thể xóa cũng như đổi mật khẩu của admin khác, root thì toàn quyền.
- API:
  - Bộ API CRUD để quản lý các file script python/golang. User chỉ có thể quyền với file của chính mình. Cho phép share file script với các user khác (user được share không có quyền xóa file)
  - API run file script sử dụng SSE để stream dữ liệu console log api về
  - API stop tiến trình chạy file script
- Build Dockerfile

Giai đoạn 2: 3 tuần

- Hoàn thành giao diện web để sử dụng tính năng
- Build Dockerfile cho cả BE và FE
