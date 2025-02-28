# GoLink

GoLink is a URL shortener service built with Go and React.

## Getting Started

1. Clone the repository:
```bash
git clone https://github.com/Okabe-Junya/golink.git
cd golink
```

2. Set up environment variables:
```bash
cp .env.example .env
```

3. Update the `.env` file with your configuration:
- Set `FIREBASE_CREDENTIALS_JSON` or `FIREBASE_CREDENTIALS_FILE` for Firebase authentication
- Adjust other variables as needed

4. Start the application:
```bash
docker compose up
```

The application will be available at:
- Frontend: http://localhost:3001
- Backend: http://localhost:8080
- Firestore Emulator: localhost:8081

## Development

### Backend Development

The backend is written in Go and uses:
- Gin web framework
- Firebase/Firestore for data storage
- Docker for containerization

To run backend tests:
```bash
cd backend
make test
```

### Frontend Development

The frontend is built with:
- React
- TypeScript
- Tailwind CSS

To run frontend tests:
```bash
cd frontend
npm install
npm test
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| FIREBASE_CREDENTIALS_JSON | Firebase credentials in JSON format | - |
| FIREBASE_CREDENTIALS_FILE | Path to Firebase credentials file | path/to/serviceAccountKey.json |
| APP_DOMAIN | Application domain | localhost |
| PORT | Backend port | 8080 |
| FRONTEND_PORT | Frontend port | 3001 |
| BACKEND_PORT | Backend port (for Docker) | 8080 |
| FIRESTORE_EMULATOR_HOST | Firestore emulator host | firestore:8081 |
| GOOGLE_CLOUD_PROJECT | GCP project ID | golink-local |

## License

This project is licensed under the [GNU AFFERO GENERAL PUBLIC LICENSE Version 3](LICENSE).

