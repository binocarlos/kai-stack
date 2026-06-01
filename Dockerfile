### API Base ###
#---------------
  FROM golang:1.25.0-alpine AS api-base
  WORKDIR /app
  COPY go.mod go.sum ./
  RUN go mod download
  
  ### API Development ###
  #----------------------
  FROM api-base AS api-dev-env
  # - ffmpeg is needed by worker job handlers (the worker now runs in-process)
  RUN apk add --no-cache bash ffmpeg
  # - Air provides hot reload for Go
  RUN go install github.com/air-verse/air@v1.52.3
  # - Copy the files and run a build to make startup faster
  COPY api /app/api
  WORKDIR /app/api
  # - Run a build to make the intial air build faster
  RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o /api
  # - Entrypoint is the air command. `serve` runs both the API and the worker
  #   in one process, controlled by SERVER_ENABLED / WORKER_ENABLED.
  ENTRYPOINT ["air", "--build.bin", "/api", "--build.cmd", "CGO_ENABLED=0 go build -ldflags \"-s -w\" -o /api", "--"]
  CMD ["serve"]


  #### API Build ###
  #-----------------------
  FROM api-base AS api-build-env
  # Following git lines required for buildvcs to work
  RUN apk add --no-cache git
  COPY .git /app/.git
  COPY api /app/api
  WORKDIR /app/api
  # - main.version is a variable required by Sentry and is set in .drone.yaml
  ARG APP_VERSION="v0.0.0+unknown"
  RUN CGO_ENABLED=0 go build -buildvcs=true -ldflags "-s -w -X main.version=$APP_VERSION -X github.com/Bayes-Price/Platinum/api/pkg/data.Version=$APP_VERSION" -o /api
  
  ### Frontend Base ###
  #--------------------
  FROM node:23-alpine AS ui-base
  WORKDIR /app
  # - Install dependencies
  COPY ./frontend/package.json /app/
  COPY ./frontend/yarn.lock /app/yarn.lock
  RUN yarn install
  
  
  ### Frontend Development ###
  #---------------------------
  FROM ui-base AS ui-dev-env
  ENTRYPOINT ["yarn"]
  CMD ["run", "dev"]
  
  ### Frontend Build ###
  #----------------------
  FROM ui-base AS ui-build-env
  # Supabase frontend config - Vite inlines VITE_* at build time. The anon key
  # is public, so embedding it in the static bundle is expected and safe.
  ARG VITE_SUPABASE_URL
  ARG VITE_SUPABASE_ANON_KEY
  ENV VITE_SUPABASE_URL=$VITE_SUPABASE_URL
  ENV VITE_SUPABASE_ANON_KEY=$VITE_SUPABASE_ANON_KEY
  # Copy the rest of the code
  COPY ./frontend /app
  # Build the frontend
  RUN yarn build
  
  # frontend production image
  FROM nginx:alpine AS frontend-final-build
  LABEL maintainer="kai@kaidam.ltd"
  COPY ./nginx.new.conf /etc/nginx/nginx.conf
  COPY --from=ui-build-env /app/dist /www
  
  # api production image — `serve` runs both the API and the worker in one
  # process, controlled by SERVER_ENABLED / WORKER_ENABLED.
  FROM golang:1.24.2-alpine AS api-final-build
  RUN apk --update add --no-cache ca-certificates ffmpeg
  COPY --from=api-build-env /api /api
  EXPOSE 80
  ENTRYPOINT ["/api", "serve"]
