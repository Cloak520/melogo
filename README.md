# MeloGo

A self-hosted music streaming server built with Go, featuring automatic music library scanning, playlist management, and a web-based interface for music playback.

## Features

- **Automatic Music Library Scanning**: Automatically scans your music directory and extracts metadata (title, artist, album, duration)
- **Web-based Music Player**: Stream your music collection through a modern web interface
- **User Authentication**: Secure login and registration system with JWT-based authentication
- **Playlist Management**: Create, edit, and manage custom playlists
- **Favorites System**: Mark your favorite songs for easy access
- **Lyrics Support**: Automatic lyrics scraping and display
- **Cover Art**: Automatic cover art extraction and display
- **Search Functionality**: Search through your music library by title, artist, or album
- **Multi-language Support**: i18n support for multiple languages
- **Admin Panel**: Administrative tools for managing music library
- **Docker Support**: Easy deployment with Docker and Docker Compose

## Technology Stack

- **Backend**: Go (Gin framework)
- **Database**: SQLite
- **Frontend**: HTML, CSS, JavaScript
- **Authentication**: JWT (JSON Web Tokens)
- **Music Metadata**: go-taglib
- **Internationalization**: go-i18n
- **Template Engine**: Go's html/template

## Installation

### Prerequisites

- Go 1.25 or higher
- Git

### Quick Start

1. Clone the repository:
   ```bash
   git clone https://github.com/your-username/MeloGo.git
   cd MeloGo
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Build the application:
   ```bash
   go build -o melogo
   ```

4. Configure the application by editing `.env` file:
   ```bash
   # Edit .env to configure your settings
   cp .env.example .env
   ```

5. Run the application:
   ```bash
   ./melogo
   ```
   Or with a custom .env file:
   ```bash
   ./melogo -f /path/to/your/.env
   ```

## Configuration

The application can be configured using environment variables in the `.env` file:

- `SERVER_HOST`: Server host (default: localhost)
- `SERVER_PORT`: Server port (default: 8080)
- `SERVER_DEBUG`: Enable debug mode (default: false)
- `DATABASE_PATH`: Path to SQLite database file (default: ./data/melogo.db)
- `MUSIC_DIRECTORY`: Directory containing music files (default: ./music)
- `MUSIC_SCAN_INTERVAL`: Music scan interval in minutes (default: 5)
- `ALLOW_REGISTRATION`: Allow user registration (default: true)
- `JWT_SECRET`: JWT secret key (change in production!)
- `LYRICS_API_URL`: API URL for lyrics scraping (default: https://api.lrc.cx)

## Usage

1. Place your music files in the configured music directory
2. Start the server - it will automatically scan and index your music
3. Access the web interface at `http://localhost:8080`
4. Register an account or log in
5. Browse, search, and play your music collection
6. Create playlists and mark favorites

## Docker Deployment

MeloGo can be easily deployed using Docker:

```bash
# Pull and run the official image
docker run -d \
  --name melogo \
  -p 8080:8080 \
  -v ./data:/home/melogo/data \
  -v ./music:/home/melogo/music \
  -e SERVER_HOST=0.0.0.0 \
  -e JWT_SECRET=your-secret-key \
  soulcloak/melogo:latest
```

Or using Docker Compose (recommended):

```yaml
version: '3.8'

services:
  melogo:
    image: soulcloak/melogo:latest
    container_name: melogo
    environment:
      - SERVER_HOST=0.0.0.0
      - SERVER_PORT=8080
      - DATABASE_PATH=./data/melogo.db
      - MUSIC_DIRECTORY=./music
      - JWT_SECRET=your-secret-key
      - ALLOW_REGISTRATION=true
      - SERVER_DEBUG=false
      - LYRICS_API_URL=https://api.lrc.cx
    ports:
      - "8080:8080"
    volumes:
      - ./data:/home/melogo/data
      - ./music:/home/melogo/music
    restart: unless-stopped
```

## API Endpoints

MeloGo provides a RESTful API at `/api/v1`:

- `POST /api/v1/register` - User registration
- `POST /api/v1/login` - User login
- `POST /api/v1/logout` - User logout
- `GET /api/v1/user/profile` - Get user profile
- `PUT /api/v1/user/profile` - Update user profile
- `GET /api/v1/songs` - List all songs
- `GET /api/v1/songs/:id` - Get song details
- `GET /api/v1/songs/:id/stream` - Stream song audio
- `GET /api/v1/songs/:id/lyrics` - Get song lyrics
- `GET /api/v1/songs/:id/cover` - Get song cover image
- `GET /api/v1/playlists` - List user playlists
- `POST /api/v1/playlists` - Create playlist
- `PUT /api/v1/playlists/:id` - Update playlist
- `DELETE /api/v1/playlists/:id` - Delete playlist
- `GET /api/v1/playlists/:id/detail` - Get playlist details
- `POST /api/v1/playlists/:id/songs` - Add song to playlist
- `DELETE /api/v1/playlists/:id/songs/:song_id` - Remove song from playlist
- `GET /api/v1/favorites` - List favorite songs
- `POST /api/v1/favorites` - Add song to favorites
- `DELETE /api/v1/favorites/:song_id` - Remove song from favorites
- `GET /api/v1/search` - Search songs
- `GET /api/v1/search/history` - Get search history

## Development

To contribute to MeloGo:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Commit your changes (`git commit -m 'Add some amazing feature'`)
5. Push to the branch (`git push origin feature/amazing-feature`)
6. Open a Pull Request

### Running in Development Mode

```bash
# Enable debug mode by setting SERVER_DEBUG=true in .env
go run main.go
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.