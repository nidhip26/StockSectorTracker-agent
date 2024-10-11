#Start the go app build
From golang:latest AS build

WORKDIR /app

COPY go.mod go.sum agent.go ./

#download dependencies
RUN go mod download

#show the contents
RUN pwd && find ./

#identify listening port
EXPOSE 8080

#start the application
CMD ["go", "run", "agent.go"]
