FROM golang:1.25-bookworm

# Install Fyne dependencies
RUN apt-get update && apt-get install -y \
    libgl1-mesa-dev \
    xorg-dev \
    libxxf86vm-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy the module files
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source
COPY . .

# Build the application
RUN go build -o pool-scanner .

CMD ["./pool-scanner"]
