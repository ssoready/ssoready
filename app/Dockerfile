FROM node:22

WORKDIR /work

COPY package.json package-lock.json ./
RUN npm ci

COPY .env tailwind.config.js build.mjs tsconfig.json ./
COPY bin ./bin
COPY public ./public
COPY src ./src

RUN npm run build-docker

ENTRYPOINT ["npm", "run", "serve-docker"]
