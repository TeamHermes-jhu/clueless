FROM node:24-alpine AS build-stage

WORKDIR /app

COPY package.json package-lock.json ./

RUN npm ci

COPY . .

ARG BACKEND_URL=http://localhost:8000
ENV BACKEND_URL=${BACKEND_URL}

RUN npm run build -- --configuration=production

FROM nginx:stable-alpine AS release-stage

RUN rm -rf /usr/share/nginx/html/*

COPY --from=build-stage /app/dist/clueless-frontend-app/browser /usr/share/nginx/html

EXPOSE 80

ENTRYPOINT ["nginx", "-g", "daemon off;"]

