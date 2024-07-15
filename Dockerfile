FROM node:18-alpine AS client-builder

# Adding pnpm and git
ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
RUN corepack enable
RUN apk add --no-cache git

# Cloning the client repository
RUN git clone https://github.com/nilotpaul/go-downloader-client /client

WORKDIR /client

RUN pnpm install

# Build will generate a bundled static folder dist
RUN pnpm build

FROM golang:1.22-alpine AS server-builder

ENV GO111MODULE=on

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

# Copying the dist directory from previous client stage
COPY --from=client-builder /client/dist ./dist

RUN go build -tags '!dev' -o bin/go-downloader

FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /root/

ENV ENVIRONMENT="PROD"
ENV PORT="3000"
ENV REDIRECT_AFTER_LOGIN="/settings"
ENV REDIRECT_AFTER_LOGOUT="/"

# Copying the dist and go binary from previous stages
COPY --from=server-builder /app/bin ./bin
COPY --from=server-builder /app/dist ./dist
COPY --from=server-builder /app/migrations ./migrations

EXPOSE 3000

CMD [ "./bin/go-downloader" ]