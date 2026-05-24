# PointSwap - NYSC Clothes Swapping App (Backend)

A high-performance RESTful API built with Go and PostgreSQL that powers the PointSwap mobile application. Handles authentication, real-time messaging, location-based product filtering, and swap verification for NYSC corps members.

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?style=flat&logo=postgresql)](https://www.postgresql.org/)
[![Gin Framework](https://img.shields.io/badge/Gin-1.9-00ADD8?style=flat)](https://gin-gonic.com/)

---

## 🎯 Features

### Core API Functionality
- **🔐 JWT Authentication** - Secure token-based authentication with middleware
- **📍 Location Services** - Haversine formula-based camp detection from GPS coordinates
- **💬 Real-Time Messaging** - Ably integration for instant messaging, typing indicators, and presence
- **🔔 Push Notifications** - Firebase Cloud Messaging for message and match notifications
- **🤝 Smart Product Matching** - Mutual matching algorithm finds perfect swap partners
- **🎫 Swap Verification** - Generates unique 6-digit codes for secure in-person swaps
- **☁️ Media Management** - Cloudinary integration for image and audio uploads

### Performance & Optimization
- **Token Caching** - Ably token caching reduces generation time from 706ms to ~5ms (140x faster!)
- **Concurrent Processing** - Goroutines for background tasks (notifications, matching)
- **Connection Pooling** - Efficient database connection management
- **Indexing** - Strategic database indexes for fast queries

---

## 🛠️ Tech Stack

### Core
- **Go 1.21+** - High-performance backend language
- **Gin** - Fast HTTP web framework
- **PostgreSQL 16** - Primary relational database
- **pq** - PostgreSQL driver for Go

### Authentication & Security
- **JWT (golang-jwt/jwt)** - JSON Web Token authentication
- **Bcrypt** - Password hashing
- **UUID** - Unique identifier generation

### Real-Time & Notifications
- **Ably Go SDK** - Real-time messaging platform
- **Firebase Cloud Messaging** - Push notifications
- **go-tokenbuilder** - Agora voice call token generation (ready, not yet used)

### Media & External Services
- **Cloudinary SDK** - Image and audio file storage
- **Google OAuth2 API** - Token verification

### Development Tools
- **godotenv** - Environment variable management
- **CORS middleware** - Cross-origin resource sharing

---

## 📊 Database Schema

### Core Tables

**users**
```sql
- user_id (UUID, PK)
- email (VARCHAR, UNIQUE)
- password (VARCHAR, hashed)
- first_name (VARCHAR)
- last_name (VARCHAR)
- avatar_url (TEXT)
- location_state (VARCHAR)
- camp_name (VARCHAR)
- latitude (DECIMAL)
- longitude (DECIMAL)
- is_online (BOOLEAN)
- last_seen (TIMESTAMP)
- created_at (TIMESTAMP)
- updated_at (TIMESTAMP)
```

**products**
```sql
- product_id (UUID, PK)
- seller_id (UUID, FK -> users)
- title (VARCHAR)
- category (VARCHAR)
- estimated_size (VARCHAR)
- status (VARCHAR) CHECK ('active', 'swapped', 'deleted')
- created_at (TIMESTAMP)
- updated_at (TIMESTAMP)
```

**product_photos**
```sql
- photo_id (UUID, PK)
- product_id (UUID, FK -> products)
- image_url (TEXT)
- display_order (INTEGER)
```

**product_wants**
```sql
- want_id (UUID, PK)
- product_id (UUID, FK -> products)
- want_user_id (UUID, FK -> users)
- wanted_category (VARCHAR)
- wanted_size (VARCHAR)
- created_at (TIMESTAMP)
```

**conversations**
```sql
- conversation_id (UUID, PK)
- created_at (TIMESTAMP)
- updated_at (TIMESTAMP)
```

**conversation_participants**
```sql
- id (UUID, PK)
- conversation_id (UUID, FK -> conversations)
- user_id (UUID, FK -> users)
- joined_at (TIMESTAMP)
```

**messages**
```sql
- id (UUID, PK)
- conversation_id (UUID, FK -> conversations)
- sender_id (UUID, FK -> users)
- message_text (TEXT)
- image_url (TEXT)
- audio_url (TEXT)
- audio_duration (INTEGER)
- is_read (BOOLEAN)
- created_at (TIMESTAMP)
```

**notifications**
```sql
- notification_id (UUID, PK)
- user_id (UUID, FK -> users)
- notification_type (VARCHAR) CHECK ('new_message', 'new_chat', 'product_inquiry', 'product_match')
- title (VARCHAR)
- message (TEXT)
- related_product_id (UUID)
- related_user_id (UUID)
- related_conversation_id (UUID)
- is_read (BOOLEAN)
- is_pushed (BOOLEAN)
- created_at (TIMESTAMP)
- read_at (TIMESTAMP)
```

**swap_requests**
```sql
- swap_id (UUID, PK)
- conversation_id (UUID, FK -> conversations)
- initiator_id (UUID, FK -> users)
- recipient_id (UUID, FK -> users)
- product_id (UUID, FK -> products)
- status (VARCHAR) CHECK ('pending', 'accepted', 'rejected', 'completed', 'cancelled')
- swap_code (VARCHAR(6))
- created_at (TIMESTAMP)
- accepted_at (TIMESTAMP)
- completed_at (TIMESTAMP)
```

**camps**
```sql
- id (UUID, PK)
- state_name (VARCHAR)
- camp_name (VARCHAR)
- latitude (DECIMAL)
- longitude (DECIMAL)
- radius_km (INTEGER) DEFAULT 10
- created_at (TIMESTAMP)
```

**push_tokens**
```sql
- id (UUID, PK)
- user_id (UUID, FK -> users)
- token (TEXT, UNIQUE)
- device_type (VARCHAR) CHECK ('ios', 'android')
- created_at (TIMESTAMP)
```

---

## 🏗️ Project Structure

```
pointswap-backend/
├── main.go                      # Application entry point
├── config/
│   └── database.go              # Database connection setup
├── models/
│   ├── user.go                  # User model & structs
│   ├── product.go               # Product model & structs
│   ├── message.go               # Message model & structs
│   ├── notification.go          # Notification model & structs
│   └── oauth.go                 # OAuth request/response structs
├── handlers/
│   ├── user_handler.go          # User auth & profile endpoints
│   ├── product_handler.go       # Product CRUD & listing
│   ├── message_handler.go       # Messaging endpoints
│   ├── notification_handler.go  # Notification endpoints
│   ├── location_handler.go      # Camp detection & location
│   ├── upload_handler.go        # Cloudinary media uploads
│   ├── swap_handler.go          # Swap request management
│   └── agora_handler.go         # Agora token generation (ready)
├── services/
│   ├── message_service.go       # Ably messaging & token caching
│   ├── notification_service.go  # Push notifications & matching
│   └── auth_service.go          # JWT generation & validation
├── repository/
│   └── message_repository.go    # Database queries for messages
├── middleware/
│   └── auth_middleware.go       # JWT verification middleware
├── routes/
│   └── routes.go                # API route definitions
├── utils/
│   ├── response.go              # Standard API response helpers
│   └── token.go                 # JWT token utilities
├── .env                         # Environment variables (not committed)
└── go.mod                       # Go module dependencies
```

---

## 🚀 Getting Started

### Prerequisites
- Go 1.21 or higher
- PostgreSQL 16+
- Ably account (for real-time messaging)
- Cloudinary account (for media storage)

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/yourusername/pointswap-backend.git
   cd pointswap-backend
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Set up PostgreSQL database**
   ```bash
   createdb pointswap
   ```

4. **Run database migrations**
   
   Execute the SQL schema from `database/schema.sql`:
   ```bash
   psql -d pointswap -f database/schema.sql
   ```

5. **Configure environment variables**
   
   Create a `.env` file:
   ```env
   # Database
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=postgres
   DB_PASSWORD=yourpassword
   DB_NAME=pointswap
   DB_SSLMODE=disable

   # JWT
   JWT_SECRET=your-super-secret-jwt-key-change-this

   # Ably
   ABLY_KEY=your-ably-api-key

   # Cloudinary
   CLOUDINARY_CLOUD_NAME=your-cloud-name
   CLOUDINARY_API_KEY=your-api-key
   CLOUDINARY_API_SECRET=your-api-secret

   # Agora (for voice calls - optional)
   AGORA_APP_ID=your-agora-app-id
   AGORA_APP_CERTIFICATE=your-agora-certificate

   # Server
   PORT=8080
   ```
6. **Run the server**
   ```bash
   go run main.go
   ```

   Server will start at `http://localhost:8080`

---

## 📡 API Endpoints

### Authentication
```
POST   /pointSwapApi/v1/register          - Register new user
POST   /pointSwapApi/v1/login             - Login with email/password
GET    /pointSwapApi/v1/userProfile       - Get current user profile
POST   /pointSwapApi/v1/profileSetUp      - Complete profile setup
PUT    /pointSwapApi/v1/users/status      - Update online status
GET    /pointSwapApi/v1/users/:id/status  - Get user status
```

### Products
```
GET    /pointSwapApi/v1/products                    - Get product feed (location-filtered)
GET    /pointSwapApi/v1/products/:product_id        - Get product by ID
POST   /pointSwapApi/v1/products                    - Create new product
PUT    /pointSwapApi/v1/products/:product_id        - Update product
DELETE /pointSwapApi/v1/products/:product_id        - Delete product
GET    /pointSwapApi/v1/my-products                 - Get current user's products
POST   /pointSwapApi/v1/products/:product_id/wants  - Set product wants
```

### Messaging
```
GET    /pointSwapApi/v1/conversations                          - Get user conversations
POST   /pointSwapApi/v1/conversations/:id/messages             - Send message
GET    /pointSwapApi/v1/conversations/:id/messages             - Get message history
PUT    /pointSwapApi/v1/conversations/:id/read                 - Mark as read
GET    /pointSwapApi/v1/messages/ably-token                    - Get Ably auth token
```

### Notifications
```
GET    /pointSwapApi/v1/notifications              - Get notifications
PUT    /pointSwapApi/v1/notifications/:id/read     - Mark notification as read
GET    /pointSwapApi/v1/notifications/unread-count - Get unread count
POST   /pointSwapApi/v1/push-token                 - Register push token
```

### Location
```
POST   /pointSwapApi/v1/location/set  - Set user location (auto-detect camp)
```

### Media Upload
```
POST   /pointSwapApi/v1/upload/image?type={profile|product|chat}  - Upload image
POST   /pointSwapApi/v1/upload/audio                               - Upload audio
```

### Swap Requests
```
POST   /pointSwapApi/v1/swaprequest/conversations/:id/swap/initiate  - Initiate swap
POST   /pointSwapApi/v1/swaprequest/swap/:id/accept                  - Accept swap
POST   /pointSwapApi/v1/swaprequest/swap/:id/reject                  - Reject swap
GET    /pointSwapApi/v1/swaprequest/conversations/:id/swap/code      - Get swap code
```

### Voice Calls (Ready, not yet implemented in frontend)
```
POST   /pointSwapApi/v1/agora/token  - Generate Agora token
```

---

## 🔑 Key Features Explained

### 1. Location-Based Camp Detection

Uses the **Haversine formula** to calculate distance between user coordinates and known NYSC camps:

```go
func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
    const R = 6371 // Earth radius in km
    
    dLat := (lat2 - lat1) * math.Pi / 180
    dLon := (lon2 - lon1) * math.Pi / 180
    
    a := math.Sin(dLat/2)*math.Sin(dLat/2) +
         math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
         math.Sin(dLon/2)*math.Sin(dLon/2)
    
    c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
    
    return R * c
}
```

- User sends GPS coordinates
- Backend queries all camps, calculates distances
- Finds nearest camp within radius (default 10km)
- Updates user's `location_state` and `camp_name`
- Product feed auto-filters by state

### 2. Ably Token Caching

**Problem:** Generating Ably tokens took 706ms per request

**Solution:** Thread-safe in-memory cache with TTL

```go
type TokenCache struct {
    mu     sync.RWMutex
    tokens map[string]*CachedToken
}

type CachedToken struct {
    Token     string
    ExpiresAt time.Time
}
```

**Results:**
- First request: 706ms (generates & caches)
- Subsequent requests: ~5ms (returns cached token)
- Tokens valid for 1 hour
- **140x performance improvement**

### 3. Product Matching Algorithm

Finds **mutual matches** where both users want what the other has:

```go
func FindAndNotifyProductMatches(newProductID uuid.UUID) {
    // Get what new product owner wants
    newProductWants := getProductWants(newProductID)
    
    // Find products that:
    // 1. Match what new owner wants (category + size)
    // 2. Owner wants what new product has
    // 3. Active status
    // 4. Different seller
    
    query := `
        SELECT p.product_id, p.seller_id, p.title
        FROM products p
        INNER JOIN product_wants pw ON p.product_id = pw.product_id
        WHERE p.category = $1 AND p.estimated_size = $2
          AND pw.wanted_category = $3 AND pw.wanted_size = $4
          AND p.seller_id != $5 AND p.status = 'active'
    `
    
    // Notify both users
    CreateProductMatchNotification(user1, user2, product)
    CreateProductMatchNotification(user2, user1, newProduct)
}
```

Runs asynchronously in goroutine after product wants are saved.

### 4. Swap Verification System

```go
func AcceptSwapRequest(swapID string) {
    // 1. Verify recipient is accepting
    verifyRecipient(swapID, currentUser)
    
    // 2. Generate random 6-digit code
    swapCode := generateSwapCode() // 100000 - 999999
    
    // 3. Update database
    db.Exec(`
        UPDATE swap_requests 
        SET status = 'accepted', swap_code = $1, accepted_at = NOW()
        WHERE swap_id = $2
    `, swapCode, swapID)
    
    // 4. Notify initiator via Ably
    publishSwapAccepted(swapID, swapCode)
    
    // 5. Return code to recipient
    return swapCode
}
```

Both users receive the same code for in-person verification.

### 5. Server-Side Status Timeout

Marks users offline if no activity for 2 minutes:

```go
func GetUserStatus(userID string) {
    var isOnline bool
    var lastSeen time.Time
    
    db.QueryRow("SELECT is_online, last_seen FROM users WHERE user_id = $1", userID)
    
    // Check timeout
    if isOnline && time.Since(lastSeen) > 2*time.Minute {
        isOnline = false
        
        // Update in background
        go db.Exec("UPDATE users SET is_online = false WHERE user_id = $1", userID)
    }
    
    return isOnline
}
```

Handles users who crash or force-close the app.

---

## 🔐 Security Features

### JWT Authentication
- Secure token generation with configurable expiry
- Middleware validates tokens on protected routes
- User context injected into request for easy access

### Password Security
- Bcrypt hashing with automatic salt
- Passwords never stored in plain text
- Never returned in API responses (json:"-" tag)

### SQL Injection Prevention
- Parameterized queries throughout
- No raw SQL string concatenation
- Input validation on all endpoints

### Rate Limiting Considerations
- Ably token caching reduces load
- Database connection pooling prevents exhaustion
- Consider adding rate limiting middleware for production

---

## 📈 Performance Optimizations

### Database Indexing
```sql
CREATE INDEX idx_products_seller ON products(seller_id);
CREATE INDEX idx_products_status ON products(status);
CREATE INDEX idx_products_location ON products(location_state);
CREATE INDEX idx_messages_conversation ON messages(conversation_id);
CREATE INDEX idx_notifications_user ON notifications(user_id);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_location_state ON users(location_state);
```

### Concurrent Processing
```go
// Background tasks run in goroutines
go func() {
    notificationService.FindAndNotifyProductMatches(productID)
}()

go func() {
    notificationService.SendMessageNotification(recipientID, message)
}()
```

### Connection Pooling
```go
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)
```

---

## 🧪 Testing

### Manual Testing with curl

**Register user:**
```bash
curl -X POST http://localhost:8080/pointSwapApi/v1/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'
```

**Login:**
```bash
curl -X POST http://localhost:8080/pointSwapApi/v1/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'
```

**Get products (with auth):**
```bash
curl -X GET http://localhost:8080/pointSwapApi/v1/products \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

---

## 🚧 Deployment

### Environment Setup

1. **Production database**
   - Set up managed PostgreSQL (e.g., AWS RDS, DigitalOcean)
   - Update connection string in `.env`

2. **Environment variables**
   - Never commit `.env` or `serviceAccountKey.json`
   - Use environment variables in production

3. **Build binary**
   ```bash
   CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .
   ```

4. **Run**
   ```bash
   ./main
   ```

### Docker (Optional)

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/.env .
COPY --from=builder /app/serviceAccountKey.json .
EXPOSE 8080
CMD ["./main"]
```

Build & run:
```bash
docker build -t pointswap-backend .
docker run -p 8080:8080 pointswap-backend
```

---

## 🐛 Known Issues & Limitations

### Current Limitations
- No rate limiting implemented (add for production)
- Push notifications tested in development only
- Agora voice calls handler ready but not integrated in frontend yet
- Only sample camps in database (need full NYSC camps list)
- No automated tests yet (consider adding unit tests)

### Production Considerations
- Add request logging middleware
- Implement rate limiting (e.g., with `golang.org/x/time/rate`)
- Set up proper logging (e.g., with `logrus` or `zap`)
- Add database backup strategy
- Implement graceful shutdown
- Add health check endpoint
- Set up monitoring (e.g., Prometheus + Grafana)

---

## 🚧 Roadmap

- [ ] Unit tests for handlers and services
- [ ] Integration tests for API endpoints
- [ ] Rate limiting middleware
- [ ] Request logging with structured logs
- [ ] Database migrations system (e.g., `golang-migrate`)
- [ ] API documentation with Swagger/OpenAPI
- [ ] Admin dashboard endpoints
- [ ] User reporting system
- [ ] Product image moderation
- [ ] Swap completion confirmation
- [ ] User ratings & reviews
- [ ] Analytics endpoints

---


### Code Style
- Follow Go best practices and conventions
- Run `go fmt` before committing
- Add comments for exported functions
- Keep functions focused and single-purpose

---

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

---

## 👨‍💻 Author

**Mukoro Oghenetega**
- LinkedIn: [linkedin.com/in/oghenetega-mukoro](https://www.linkedin.com/in/oghenetega-mukoro)
- Email: tmukoro62@gmail.com

---


**Built with 🚀 for scalable, real-time communication**

---

_Note: This is the backend repository. For the frontend (React Native), see [pointswap-app](https://github.com/Tmukoro/PointSwap-App)_
