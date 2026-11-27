# Sổ tay tự học DevOps

Ghi lại các khái niệm học được qua từng phase của dự án. Mỗi mục trả lời: **là gì, tại sao cần, áp dụng ở đâu trong dự án**.
Cuối mỗi phase có phần **Câu hỏi phỏng vấn** để tự kiểm tra.

---

## Phase 0: Tài khoản AWS & công cụ

### Root user và IAM user
- **Root** là tài khoản tạo bằng email đăng ký, có toàn quyền, kể cả đóng tài khoản hay đổi thanh toán. Không giới hạn được quyền của root.
- Chỉ dùng root cho: bật MFA, bật IAM access to Billing, việc liên quan thanh toán. Hằng ngày dùng IAM user.
- Dự án dùng IAM user `loganlee-devops`, có bật MFA và **không có access key**.

### IAM Identity Center và IAM user
- **IAM Identity Center** dùng cho công ty có nhiều người, nhiều tài khoản AWS. Mô hình: `User + Permission set + Account`. Cấp credentials tạm thời. AWS khuyên dùng cho tổ chức.
- **IAM user** gắn với 1 tài khoản. Một người làm thì đủ, với điều kiện có MFA và không có access key.
- Single-region và multi-region: multi-region là sao chép Identity Center sang region khác để dự phòng (disaster recovery). Cá nhân thì không cần.

### Group và policy
- Gán quyền cho **group** (theo vai trò: `Admins`, `Developers`, `ReadOnly`), không gán trực tiếp cho từng user. Người vào, người nghỉ thì chỉ cần thêm hoặc bớt khỏi group.
- Đặt tên group theo vai trò, không theo tên người.

### MFA
- Lộ mật khẩu mà không có điện thoại thì vẫn không đăng nhập được.
- Đặt tên thiết bị theo máy và app (`iphone-google-auth`) để biết gỡ cái nào khi mất máy.
- Lưu secret key của MFA vào password manager để khôi phục.

### Access key và credentials tạm thời
- **Access key** không tự hết hạn. Lộ lên GitHub thì vài phút sau đã có bot dùng để đào coin.
- Thay bằng:
  - Máy cá nhân: `aws login --profile todo` lấy credentials tạm thời từ đăng nhập Console.
  - CI (Phase 5): **OIDC**. GitHub Actions xin credentials tạm thời, không lưu secret.
- Nguyên tắc: **least privilege** + **không có credentials dài hạn**.

### Billing access cho IAM user
- Mặc định IAM user, kể cả có `AdministratorAccess`, **không xem được Billing**. Root phải bật *Activate IAM Access*. Đây là một "bẫy" nổi tiếng.

### FinOps: Budgets và Cost Anomaly Detection
- **Budget** báo khi chạm ngưỡng cố định ($10/$20/$50).
- **Anomaly Detection** dùng ML học mức tiêu bình thường rồi báo khi bất thường. Gửi email chỉ có ở chế độ Daily/Weekly summary. Muốn báo ngay (Individual) thì phải qua SNS.
- **SNS** là dịch vụ pub/sub: 1 topic, nhiều nơi đăng ký nhận (email, SMS, Lambda…).
- **ARN** là địa chỉ duy nhất của tài nguyên AWS: `arn:aws:<service>:<region>:<account-id>:<resource>`.

### Free tier 2025+
- Gói mới: $100 khi đăng ký + tối đa $100 khi làm 5 activity (mỗi cái $20). Tài khoản dự án đang ở **Paid plan**, còn $99.99.

### Câu hỏi phỏng vấn
1. Vì sao không dùng root hằng ngày?
2. Access key khác credentials tạm thời thế nào? OIDC giải quyết vấn đề gì?
3. Vì sao gán policy cho group thay vì user?
4. Có quyền Admin mà vẫn không vào được Billing, vì sao?
5. Budget khác Cost Anomaly Detection thế nào?

---

## Phase 1: App tối giản (viết để dễ vận hành)

### Git flow
- `main` luôn ổn định. Mỗi phase làm trên branch riêng (`feat/phase-1-app`), sau đó mở PR, chạy CI/review, rồi merge. Đây là **GitHub Flow**.
- Commit theo **Conventional Commits**: `feat(todo-service): add crud api`. Một dòng, ngắn, viết thường.

### Scaffold bằng công cụ chính thức
- `npm create vite`, `fastify-cli generate`, `go mod init`, `helm create`… rồi dọn file demo.
- Template có thể lỗi thời: fastify-cli chưa theo kịp TypeScript 6 (phải thêm `rootDir` và `"types": ["node"]`). Đọc kỹ lỗi là sửa được.

### Vì sao Go cho todo-service
- Docker, Kubernetes, Terraform, Prometheus, ArgoCD đều viết bằng Go.
- Biên dịch ra 1 file binary nên image nhỏ (khoảng 15MB, Node từ 150MB). Pull nhanh, ít lỗ hổng.
- Go + Node giống thực tế nhiều ngôn ngữ, CI phải xử lý nhiều kiểu build.

### 12-Factor App: config qua biến môi trường
- `DATABASE_URL`, `PORT` đọc từ env. Cùng một image chạy ở mọi môi trường.
- Local truyền qua docker-compose, K8s qua Secret, AWS qua External Secrets ← Secrets Manager.
- Không hardcode mật khẩu trong code.

### Liveness và readiness
| | `/healthz` (liveness) | `/readyz` (readiness) |
|---|---|---|
| Kiểm tra | Process còn sống | Dependency (DB) kết nối được |
| K8s làm gì khi fail | Restart pod | Ngừng gửi traffic, không restart |

- Không để liveness kiểm tra DB: DB chết thì K8s sẽ restart mọi pod liên tục, mà restart cũng không chữa được DB.

### Graceful shutdown
- K8s gửi **SIGTERM**, đợi `terminationGracePeriodSeconds` (mặc định 30s) rồi mới **SIGKILL**.
- App phải bắt SIGTERM, ngừng nhận request mới, xử lý nốt request đang chạy rồi thoát với exit code 0.
- Không làm thì request đang chạy bị cắt và người dùng gặp lỗi 502 khi deploy. Cần cho zero-downtime deploy.

### Metrics (Prometheus)
- `http_requests_total{method,route,status}` là counter, dùng tính traffic và tỉ lệ lỗi.
- `http_request_duration_seconds` là histogram, dùng tính latency p95/p99.
- **Golden Signals**: latency, traffic, errors, saturation.
- **High cardinality**: label phải là route pattern `/todos/{id}`, không phải `/todos/42`. Nếu không, mỗi id thành 1 time series và Prometheus hết RAM.

### Log có cấu trúc (JSON)
- Log JSON cho phép Loki lọc `status >= 500`, `duration_ms > 1000`.
- Không log `/healthz`, `/readyz`, `/metrics` vì bị gọi liên tục, chỉ thành rác và tốn tiền lưu trữ.

### Store interface + Memory store
- Handler chỉ biết interface `Store`, không biết Postgres. Unit test dùng Memory store nên chạy khoảng 1s mà không cần DB, và CI nhanh.
- `go test -race` phát hiện lỗi truy cập đồng thời (race condition).

### user-service (Node/Fastify): cùng khái niệm, khác ngôn ngữ
| Khái niệm | Go (todo-service) | Node (user-service) |
|---|---|---|
| Log JSON | `slog.NewJSONHandler` | pino (có sẵn trong Fastify) |
| Metrics | `client_golang` | `prom-client` (cùng tên metric để dùng chung dashboard) |
| Graceful shutdown | `signal.NotifyContext` + `srv.Shutdown` | `close-with-grace` + `app.close()` |
| Validation | Tự kiểm tra trong handler | JSON Schema, Fastify trả 400 trước khi vào handler |
| Bỏ log probe | Lọc trong middleware | `logLevel: 'silent'` trên route |

- **Tự viết entrypoint `server.ts`** thay vì `fastify start`: kiểm soát rõ việc bắt SIGTERM, đọc env, listen. Container nên chạy thẳng `node dist/server.js`.
- **Listen `0.0.0.0` chứ không phải `localhost`**: trong container, `localhost` chỉ là chính container đó. Listen localhost thì bên ngoài (Docker, K8s Service) không gọi vào được. Đây là lỗi kinh điển khi mới đóng gói Docker.
- **HTTP status đúng nghĩa**: 201 Created, 204 No Content, 400 dữ liệu sai, 404 không tìm thấy, **409 Conflict** khi trùng email (DB báo mã lỗi `23505` unique violation), 503 khi chưa sẵn sàng.
- **Đặt tên metric giống nhau giữa các service** (`http_requests_total`) thì 1 dashboard Grafana dùng được cho mọi service, chỉ cần lọc theo label `job`.

### Bài học debug
- zsh **không tự tách biến** như bash: `J='-H a:b'; curl $J` truyền cả chuỗi thành 1 tham số. Header bị dấu cách thừa nên Node trả 400 "Client Error" ngay ở tầng parser HTTP, request không vào tới router (metrics không đếm).
- Cách debug: `curl -v` xem request thô. Request không xuất hiện trong metrics/log của app nghĩa là bị chặn trước khi tới code.
- Template/thư viện đổi API theo phiên bản: Fastify 5 coi `err` trong error handler là `unknown`, phải khai báo `FastifyError`.

### Frontend gọi API qua reverse proxy
```
Browser → /api/todos/* → proxy → todo-service /todos/*
        → /api/users/* → proxy → user-service /users/*
```
- **CORS**: trình duyệt chặn trang ở origin A (`localhost:5173`) gọi sang origin B (`localhost:8080`) trừ khi B cho phép. Gom mọi thứ về cùng origin qua proxy thì không cần cấu hình CORS.
- Frontend chỉ dùng **đường dẫn tương đối** `/api/...`, không biết địa chỉ service. Proxy lo việc định tuyến:
  - Dev: Vite `server.proxy`
  - Docker (Phase 2): nginx
  - K8s (Phase 4–6): Ingress / AWS ALB
- Cùng một bản build frontend chạy ở mọi môi trường. Đây là 12-Factor áp dụng cho frontend.
- Proxy **bỏ tiền tố `/api`** (path rewrite) nên service không cần biết mình đứng sau proxy.

### Bài học debug: xung đột cổng
- Gọi `localhost:5173` mà nhận về trang của app khác: cổng đã bị process khác chiếm, Vite với `--strictPort` không mở được.
- Kiểm tra: `lsof -nP -iTCP:5173 -sTCP:LISTEN`. Đổi cổng (`--port 15173`) thay vì tắt process của người khác.
- Trong K8s không còn chuyện này: mỗi pod có IP riêng, mọi container cùng listen 8080 vẫn được.

### Giới hạn đã biết (để cải tiến sau)
- Tự tạo bảng bằng `CREATE TABLE IF NOT EXISTS` lúc khởi động. Thực tế dùng công cụ migration (golang-migrate, Flyway) chạy bằng K8s Job hoặc initContainer.

### Câu hỏi phỏng vấn
1. Liveness khác readiness thế nào? Nếu liveness kiểm tra DB thì chuyện gì xảy ra?
2. Graceful shutdown là gì, liên quan gì đến SIGTERM trong Kubernetes?
3. High cardinality trong Prometheus là gì, tránh thế nào?
4. Vì sao log nên ở dạng JSON?
5. 12-Factor App nói gì về config?
6. Vì sao app trong container phải listen `0.0.0.0`?
7. Khi nào trả 400, 404, 409, 503?
8. CORS là gì? Reverse proxy giải quyết nó thế nào?
9. Vì sao frontend nên gọi API bằng đường dẫn tương đối?
