# 🧠 CueMind – Intelligent Flashcard Generator

CueMind is a modern Anki-style spaced repetition app powered by LLMs and Go.  
Users can upload learning materials (PDFs, PPTX, DOCX), and the system automatically generates high-quality flashcards using Google's Gemini API. Flashcards are reviewed through a dynamic interface that adapts based on user memory strength.

---

## 🚀 Features

- ✨ **LLM-Powered Flashcards**: Upload files, and Gemini generates flashcards in JSON format
- 🔁 **Spaced Repetition Engine**: Implements a simplified SM-2 algorithm for personalized reviews
- 📁 **Smart File Uploads**: Upload any file (PDF, PPTX, DOCX). Files are converted and stored in S3 via pre-signed URLs
- 🧵 **Asynchronous Processing**: Background workers consume jobs from RabbitMQ and handle heavy LLM requests
- 🔐 **JWT Authentication**: Secure login and protected routes
- 🖥️ **Real-Time Feedback**: Users are notified via WebSockets when new cards are ready
- 🌱 **Clean Project Structure**: Modular Go services for DB, LLM, file processing, and WebSocket management

---

## 🛠 Tech Stack

| Layer          | Tech                                         |
|----------------|----------------------------------------------|
| Backend        | Go (Chi + net/http)                          |
| Frontend       | React (with hooks)                           |
| LLM            | Gemini (Google's Generative AI)              |
| Storage        | AWS S3 (with pre-signed URL uploads)         |
| Messaging Queue| RabbitMQ                                     |
| Real-Time      | WebSockets (Gorilla WebSocket)               |
| DB             | PostgreSQL (with sqlc)                       |
| Auth           | JWT-based authentication                     |
| File Conversion| Custom microservice (LibreOffice or unioffice) |

---

## 📦 Project Structure

```
├── cmd/                # Main entry points (server, worker)
├── internal/
│   ├── api/            # HTTP handlers
│   ├── auth/           # JWT logic
│   ├── database/       # SQLC queries + models
│   ├── llm/            # Gemini client and flashcard generation
│   ├── storage/        # S3 uploader + presigned logic
│   ├── queue/          # RabbitMQ publisher/consumer
│   ├── converter/      # File conversion service (pptx/docx → pdf)
│   └── ws/             # WebSocket hub for real-time feedback
├── sql/                # DB migrations and queries
└── frontend/           # React SPA (study interface)
```

---

## 🔁 How It Works

1. **User uploads a file** → stored in S3
2. **Server queues a job** → RabbitMQ task created
3. **Worker picks the job** → downloads + converts file → sends to Gemini → receives flashcards
4. **Flashcards saved to DB**
5. **WebSocket notifies client** → user begins review

---

## 🧪 Local Development

1. Run PostgreSQL and RabbitMQ using Docker
2. Start backend: `go run cmd/server/main.go`
3. Start worker: `go run cmd/worker/main.go`
4. Start frontend: `npm start` (React app)
5. Backend API: `localhost:8000`, Frontend: `localhost:3000`

---

## 💡 Future Ideas

- [ ] Track spaced repetition stats per user
- [ ] Add retry logic & dead-letter queue for failed jobs
- [ ] Deploy with Fly.io / Vercel / Railway
- [ ] Export flashcards as Anki .apkg format

---

## 🙌 Acknowledgements

- Inspired by Anki and Habitica
- Built using Google Gemini, AWS S3, and PostgreSQL
- Demonstrates full-stack architecture, messaging, file conversion, and AI integration

---