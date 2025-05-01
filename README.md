🧠 CueMind – Intelligent Flashcard Generator

CueMind is a modern Anki-style spaced repetition app powered by LLMs and Go.
Users can upload learning materials (PDFs), and the system automatically generates high-quality flashcards using Google's Gemini API. Flashcards are reviewed through a dynamic interface that adapts based on user memory strength.
🚀 Features

    ✨ LLM-Powered Flashcards: Upload PDFs, and Gemini generates flashcards in JSON format

    🗃️ Spaced Repetition Engine: Implements a simplified SM-2 algorithm for personalized reviews

    📁 Smart File Uploads: Users upload via pre-signed AWS S3 URLs (secure, scalable)

    🧵 Asynchronous Processing: Background workers consume jobs from RabbitMQ and handle heavy LLM requests

    🔐 JWT Authentication: Users are authenticated with secure token-based login

    🖥️ Real-Time Feedback: Users are notified via WebSockets when new cards are ready

    🌱 Clean Project Structure: Modular Go services for database, LLM, queue, WebSocket hub

🛠 Tech Stack
Layer	Tech
Backend	Go (Chi + net/http)
Frontend	React (with hooks, no framework)
LLM	Gemini (Google's Generative AI)
Storage	AWS S3 (pre-signed URL uploads)
Messaging Queue	RabbitMQ
Real-Time	WebSockets (Gorilla WebSocket)
DB	PostgreSQL (with sqlc)
Auth	JWT-based authentication
📦 Project Structure

├── cmd/                # Main entry points (server, worker)
├── internal/
│   ├── api/            # HTTP handlers
│   ├── auth/           # JWT logic
│   ├── database/       # SQLC queries + models
│   ├── llm/            # Gemini client and card generation
│   ├── storage/        # S3 uploader + presigned logic
│   ├── queue/          # RabbitMQ publisher/consumer
│   └── ws/             # WebSocket hub
├── sql/                # DB migrations and queries
└── frontend/           # React SPA (study interface)

🔁 How It Works

    User uploads a file → stored in S3

    Server queues a job → sent to RabbitMQ

    Worker picks the job → downloads file, sends to Gemini, receives flashcards

    Flashcards saved to DB

    WebSocket notifies client → study session begins

🧪 Local Development

    Run PostgreSQL and RabbitMQ (via Docker)

    Run go run cmd/server/main.go (starts API + WebSocket)

    Run go run cmd/worker/main.go (starts background processing)

    Frontend runs on localhost:3000, backend on localhost:8000

💡 Future Ideas

Support .pptx and .docx with built-in file conversion

Add Anki-style answer grading and statistics

Deploy to Fly.io + Vercel

    Add multi-user spaced repetition tracking

🙌 Acknowledgements

    Inspired by Anki and Habitica

    Flashcards generated with Google's Gemini API

    Built for portfolio-level demonstration of full-stack system design