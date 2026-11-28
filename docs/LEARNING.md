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

---

## Phase 2: Docker

### Multi-stage build
- **Stage build**: image to, có đủ compiler và devDependencies để biên dịch.
- **Stage runtime**: chỉ copy kết quả (binary, `dist/`, `node_modules` production). Không có compiler, không có source.
- Kết quả trong dự án:

| Image | Base runtime | Kích thước |
|---|---|---|
| todo-service (Go) | `distroless/static-debian13` | ~15MB |
| frontend | `nginx-unprivileged:1.31-alpine` | ~73MB |
| user-service (Node) | `distroless/nodejs24-debian13` | ~174MB |

### Layer cache
- Docker cache theo từng lệnh. Copy `go.mod`/`package.json` rồi cài dependency **trước**, copy source **sau**. Sửa code thì không phải tải lại dependency.
- `RUN --mount=type=cache` (BuildKit) giữ cache của Go/npm giữa các lần build.

### Distroless và non-root
- **Distroless**: không shell, không package manager, không curl. Kẻ tấn công vào được container cũng không có công cụ gì để dùng, và ít gói thì ít CVE.
- **Non-root** (`USER nonroot`, `USER 101`): container bị chiếm cũng không có quyền root. K8s có thể bắt buộc điều này (`runAsNonRoot`, Kyverno ở Phase 8).
- nginx non-root phải listen **8080** vì cổng dưới 1024 cần quyền root.
- Healthcheck trong distroless: không có curl nên dùng chính binary (`/todo-service healthcheck`) hoặc `node -e "fetch(...)"`.

### .dockerignore
- Loại `node_modules`, `dist`, test, `.env` khỏi build context: build nhanh hơn, image sạch, **không lọt secret vào image**.

### Config nginx bằng biến môi trường
- Image nginx tự chạy `envsubst` cho `/etc/nginx/templates/*.template` lúc khởi động. Địa chỉ service (`TODO_API_URL`) là env, cùng một image dùng cho compose và K8s.
- SPA fallback: `try_files $uri /index.html`. File trong `/assets/` có hash trong tên nên cache 1 năm, còn `index.html` thì `no-cache`.

### Docker Compose
- **Network nội bộ**: các service gọi nhau bằng tên (`postgres`, `todo-service`). Docker có DNS nội bộ.
- **Chỉ publish frontend** (`8000:8080`). Service và DB không mở ra host, giống mô hình K8s sau này.
- `depends_on: condition: service_healthy`: chờ DB healthy mới khởi động service.
- `restart: unless-stopped`: container chết thì tự khởi động lại.
- **Volume** `pgdata`: dữ liệu còn nguyên sau khi restart container. Postgres 18 mount ở `/var/lib/postgresql` (khác các bản cũ).
- **Database per service**: `todos` và `users` là 2 database riêng. Service không đọc bảng của nhau.

### Graceful shutdown trong container
- `docker stop` gửi SIGTERM, chờ timeout rồi SIGKILL.
- Exit code **0** là thoát đúng cách, **137** (128+9) là bị SIGKILL vì không xử lý SIGTERM. Cả 3 service đều trả 0.

### Quét lỗ hổng image (Trivy)
- `trivy image --severity HIGH,CRITICAL <image>`.
- Thực tế gặp phải: `nginx-unprivileged:1.29-alpine` có 56 HIGH, 1 CRITICAL, toàn bộ ở thư viện OS (curl, libexpat…), **không phải code của mình**.
- Cách xử lý theo thứ tự:
  1. Dùng base image mới hơn.
  2. `apk upgrade` lúc build để lấy bản vá mới nhất.
  3. Gỡ gói không dùng (`apk del curl`) để giảm bề mặt tấn công.
  4. Lỗi nằm ở chính nginx thì phải lên bản có fix (1.31 mainline). nginx.org khuyên dùng mainline.
- Kết quả: frontend 0/0, todo-service 0/0. user-service còn 2 HIGH ở `libssl3t64` **chưa có bản vá**. CI dùng `--ignore-unfixed` để không chặn build vì lỗi chưa ai sửa được.
- nginx đánh số: **chẵn = stable** (1.30), **lẻ = mainline** (1.31).

### pre-commit + gitleaks (shift left)
- Hook chạy mỗi lần `git commit`: gitleaks, detect-private-key, trailing-whitespace, end-of-file-fixer, check-yaml/json, gofmt.
- Đã thử: commit một GitHub token giả thì gitleaks chặn (`RuleID: github-pat`).
- Lỡ commit secret thì: **đổi secret ngay** (coi như đã lộ), rồi mới xóa khỏi lịch sử git.
- `tsconfig.json` cho phép comment (JSONC) nên phải loại khỏi `check-json`.
- pre-commit chỉ bảo vệ máy đã chạy `pre-commit install`. CI (Phase 3) chạy lại gitleaks để chắc chắn.

### Câu hỏi phỏng vấn
1. Multi-stage build là gì, lợi ích?
2. Vì sao copy `package.json` trước rồi mới copy source?
3. Distroless là gì? Không có shell thì debug và healthcheck thế nào?
4. Vì sao container không nên chạy bằng root?
5. Exit code 137 nghĩa là gì?
6. `depends_on` có đảm bảo service phụ thuộc đã sẵn sàng không? (Chỉ khi dùng `condition: service_healthy`.)
7. Trivy báo lỗ hổng ở base image thì xử lý ra sao? Lỗ hổng chưa có bản vá thì sao?
8. Lỡ commit secret lên GitHub thì làm gì đầu tiên?

---

## Phase 3: CI với GitHub Actions

### Cấu trúc một workflow
```
workflow (.github/workflows/*.yml)
├── on:           khi nào chạy (push, pull_request, schedule, workflow_call…)
├── permissions:  quyền của GITHUB_TOKEN
└── jobs:         mặc định chạy song song
    └── <job>:
        ├── runs-on: ubuntu-latest   (máy ảo mới mỗi lần, chạy xong bị xóa)
        ├── needs: <job khác>        (nối job thành pipeline)
        └── steps:                   (chạy tuần tự)
            ├── uses: owner/action@sha   (action làm sẵn, tham số qua with:)
            └── run: lệnh shell          (run: | cho nhiều dòng)
```
- `${{ ... }}` là biểu thức: `github.ref`, `github.sha`, `inputs.x`, `steps.<id>.outputs.y`, `secrets.GITHUB_TOKEN`.
- **Exit code khác 0 thì step fail**, kéo theo job fail và PR hiện ✗. Mọi kiểm tra trong CI đều dựa trên quy tắc này.
- `if: github.ref == 'refs/heads/main'`: step chỉ chạy ở main. Ở PR thì bị skipped (–).

### Path filter (monorepo)
- `on.push.paths` / `on.pull_request.paths`: chỉ chạy khi thư mục của app thay đổi, tiết kiệm phút CI.
- Với `pull_request`, GitHub so **toàn bộ diff của PR với base**, không chỉ commit vừa push.
- Phải thêm cả file workflow và file reusable vào `paths`.

### Reusable workflow
- File `docker-image.yml` có `on: workflow_call` + `inputs` (giống tham số hàm). Bên gọi dùng `uses: ./.github/workflows/docker-image.yml` + `with:` ở **cấp job**.
- Bên gọi phải cấp `permissions` (`packages: write`). Workflow được gọi không thể có nhiều quyền hơn bên gọi.
- DRY: 3 app dùng chung 1 job build/scan/push. Sửa một chỗ là áp dụng cho cả ba.

### Pipeline image
1. `docker/metadata-action` sinh tag `sha-<commit>` (luôn có) và `latest` (chỉ ở main). Tag theo SHA giúp truy được image build từ commit nào và rollback dễ.
2. Build với `load: true`: nạp vào Docker trên máy CI, chưa push.
3. **Trivy scan là security gate** (`exit-code: 1`, `ignore-unfixed`). Quét xong mới push, nên image có lỗ hổng không bao giờ lên registry.
4. Push lên GHCR **chỉ từ main**, đăng nhập bằng `GITHUB_TOKEN` có sẵn (không tạo secret).
- `cache-from/cache-to: type=gha`: layer cache lưu ở GitHub Actions cache.

### Bảo mật CI
- `permissions: contents: read` ở đầu file: least privilege cho `GITHUB_TOKEN`. Chỉ job cần mới được cấp thêm quyền.
- **Pin action theo commit SHA** (`@3d3c42e… # v7.0.1`): tag có thể bị kẻ xấu chuyển sang code độc hại (tấn công chuỗi cung ứng), SHA thì không đổi được.
- `concurrency` + `cancel-in-progress`: hủy run cũ khi có push mới.
- gitleaks với `fetch-depth: 0`: quét **toàn bộ lịch sử**. Secret đã xóa ở commit sau vẫn còn trong lịch sử git.
- **Defense in depth**: pre-commit trên máy (lớp 1) + gitleaks trong CI (lớp 2, áp dụng cho mọi người).
- actionlint: lint chính file workflow. Chạy trên máy (`actionlint`) và trong CI (Docker image có phiên bản cố định).

### Sự cố: PR xanh nhưng main đỏ (Trivy)
- Triệu chứng: merge PR #4 xong, cả 3 job image trên `main` fail ở bước Trivy, dù trên PR đều xanh.
- Đọc log (`gh run view <id> --log-failed`) thì thấy `unable to find 'v0.70.0'`: lỗi ở **bước cài Trivy**, không phải image có lỗ hổng.
- Nguyên nhân: trên PR, binary Trivy được lấy từ **cache**. Cache của GitHub Actions **tách theo branch**: `main` không đọc được cache của branch PR, nên phải tải mới, và installer của `trivy-action` không còn tải được.
- Chạy lại (`gh run rerun --failed`) vẫn lỗi y hệt, nên đây là lỗi cố định, không phải flaky.
- Cách sửa: chạy Trivy bằng Docker image chính thức có phiên bản cố định (`aquasec/trivy:0.76.2`), đọc image qua Docker socket. Thử trên máy trước, rồi sửa qua PR (main đã được bảo vệ).
- **Pin image theo digest** (`image:tag@sha256:...`): tag có thể bị push đè (vô tình hoặc bị tấn công supply chain), còn digest là hash nội dung, không đổi được. Giống pin action theo commit SHA. Lấy digest của bản đa kiến trúc bằng `docker buildx imagetools inspect <image>` (dòng `Digest` đầu tiên). Muốn nâng version thì cập nhật cả tag lẫn digest.
- Bài học:
  - Cache có thể **che giấu lỗi**. Bước nào cần tải từ Internet thì nên có đường không phụ thuộc cache.
  - Đọc log từ dưới lên, tìm dòng `error`/`crit` đầu tiên.
  - Ít phụ thuộc action bên thứ ba hơn thì ít rủi ro hơn. Công cụ có Docker image chính thức thì chạy thẳng image.

### Node trong CI
- `npm ci` chứ không phải `npm install`: cài đúng y `package-lock.json`, lệch là báo lỗi. CI phải **reproducible**.
- `setup-node` với `cache: npm`: dùng lock file làm khóa cache.
- `setup-go` với `go-version-file: go.mod`: máy dev và CI cùng một phiên bản Go.

### Branch protection (Ruleset)
- Luật cho `main`: bắt buộc PR, bắt buộc check `gitleaks` + `actionlint` xanh, cấm force push, cấm xóa, chỉ cho squash merge.
- **Bẫy**: đặt check có path filter (vd `todo-service / test`) là bắt buộc thì PR chỉ sửa frontend sẽ **kẹt mãi** ở "Waiting for status", vì workflow không chạy nên check không bao giờ có kết quả. Chỉ bắt buộc check **luôn chạy**. Cải tiến sau: dùng job "gate" tổng hợp kết quả.
- Repo một người thì số người duyệt bằng 0 (GitHub không cho tự approve PR của mình). Công ty thường 1–2.

### Git: danh tính và lịch sử
- Không cấu hình `user.email` thì git tự ghép `user@hostname`. Hostname là IP LAN, đổi theo wifi, nên GitHub thấy nhiều "người lạ", commit không có avatar và không tính vào contribution graph.
- Dùng email noreply `<id>+<login>@users.noreply.github.com`: gắn đúng tài khoản mà không lộ email thật. Cấu hình **local** trong repo nếu máy còn dùng cho việc khác.
- GitHub settings → Emails → "Keep my email addresses private" để squash merge trên web cũng dùng noreply.
- Viết lại lịch sử (`filter-branch`, `rebase`) rồi `push --force-with-lease`: chỉ an toàn khi chưa ai khác dùng branch. Ở công ty gần như cấm trên main. Ruleset `non_fast_forward` chặn hẳn.
- `--force-with-lease` an toàn hơn `--force`: từ chối nếu remote có commit mà mình chưa có.
- `git rebase --onto main <commit-cũ> <branch>`: chuyển branch sang base mới sau khi PR trước được squash.
- **Đổi tên head branch của PR thì GitHub đóng PR**, phải mở PR mới.
- Squash merge: cấu hình repo lấy tiêu đề và mô tả PR làm commit message, để lịch sử main gọn.

### Câu hỏi phỏng vấn
1. Cấu trúc của một GitHub Actions workflow? Job và step khác nhau thế nào?
2. Làm sao để workflow chỉ chạy khi một thư mục thay đổi? Bẫy gì khi kết hợp với required checks?
3. Reusable workflow là gì, vì sao dùng?
4. Vì sao pin action theo SHA thay vì tag?
5. Vì sao scan image trước khi push? `ignore-unfixed` để làm gì?
6. `npm ci` khác `npm install`?
7. `GITHUB_TOKEN` là gì? Vì sao đặt `permissions: contents: read`?
8. Vì sao gitleaks trong CI cần `fetch-depth: 0`?
9. `--force-with-lease` khác `--force`? Khi nào được force push?
