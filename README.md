# Gooo

PWA en Go para consumir eventos de TikTok LIVE/TikTools y reproducir chat mediante TTS con enfoque en baja latencia y resiliencia realista.

> ⚠️ **Importante**: ninguna conexión externa puede garantizar “nunca desconectarse”. Esta app aplica reconexión automática, heartbeats, backoff exponencial y deduplicación para recuperar servicio rápidamente, pero no promete disponibilidad absoluta.

## Características

- Backend Go ligero, fácil de correr y desplegar.
- PWA instalable (`manifest.webmanifest`, Service Worker, app shell cacheado).
- Eventos en tiempo real por SSE (`/events`) con estado de conexión.
- Modo `demo` funcional sin credenciales externas.
- Arquitectura de adaptadores desacoplados para TikTok LIVE y TikTools (interfaces + configuración).
- Webhook autenticado para TikTools con HMAC SHA-256, validación y límites de tamaño/rate.
- Cola acotada con prioridad + backpressure para evitar crecimiento ilimitado.
- Reconexión automática con backoff exponencial + jitter, heartbeat timeout y límite de reintentos configurable.
- Deduplicación de eventos tras reconexión.
- Controles TTS: habilitar audio, pausar/reanudar, saltar, vaciar cola, volumen, velocidad, tono, voz, idioma.
- Filtros anti-spam/comandos/usuarios bloqueados/tipos bloqueados/longitud máxima.
- Endpoints `/healthz`, `/readyz`, `/metrics`.

## Arquitectura

- `cmd/server`: entrada del servidor HTTP.
- `internal/app`: orquestación, estado, SSE, handlers, reconexión.
- `internal/connector`: interfaz de conectores y conector demo.
- `internal/events`: tipos y deduplicación.
- `internal/queue`: cola acotada con prioridad y backpressure.
- `internal/backoff`: backoff exponencial con jitter.
- `internal/webhook`: firma HMAC y rate limiting.
- `internal/tts`: filtros de mensajes para voz.
- `web/static`: UI responsive en español + PWA.

## Ejecución local

```bash
cp .env.example .env
make run
# abrir http://localhost:8080
```

También funciona con:

```bash
go run ./cmd/server
```

## Configuración

Revisa `.env.example`. Variables principales:

- `APP_MODE=demo` para iniciar sin credenciales.
- `TIKTOK_CHANNEL` canal objetivo.
- `WEBHOOK_SECRET` y `WEBHOOK_SIGNATURE_HEADER` para webhook TikTools.
- `RECONNECT_*` y `HEARTBEAT_TIMEOUT` para resiliencia.
- `QUEUE_SIZE` y `MAX_PAYLOAD_BYTES` para límites de recursos.

## Webhook TikTools

Endpoint: `POST /webhook/tiktools`

- Requiere firma `sha256=<hex>` en header configurable (`X-Tiktools-Signature` por defecto).
- Firma HMAC SHA-256 sobre el body crudo con `WEBHOOK_SECRET`.
- Límite de payload y rate limiting por origen.

Ejemplo body JSON:

```json
{
  "id": "evt-123",
  "type": "chat",
  "user": "alice",
  "text": "hola chat",
  "priority": 1
}
```

## Integración TikTok LIVE / TikTools

Este repositorio **no inventa APIs de terceros**. Se incluyen:

- Interfaces/adaptadores configurables (`internal/connector`).
- Modo demo completamente funcional para pruebas sin servicios externos.

Para integración real debes proveer endpoint/protocolo/credenciales válidas respetando términos de TikTok y TikTools. Conectores no oficiales pueden romperse con cambios externos.

## PWA y continuidad en segundo plano

- Instalable cuando el navegador emite `beforeinstallprompt`.
- App shell cacheado para UI offline.
- TTS local usando Web Speech API.
- Usa Media Session y detección de visibilidad para intentar continuidad.

> En móviles, el SO/navegador puede suspender JavaScript o voz con pantalla apagada. Se reanuda al volver cuando es posible.

## Docker

```bash
docker build -t gooo .
docker run --rm -p 8080:8080 --env-file .env gooo
```

## Calidad

```bash
make fmt
make vet
make test
make build
```

Se espera que pasen `go test ./...` y `go build ./...`.

## Troubleshooting rápido

- Estado `reconectando`: revisar endpoint/red externa; en modo demo es normal por caídas simuladas.
- Sin audio: pulsar **Habilitar audio** (política autoplay).
- Mensajes no se oyen: revisar filtros (`!comandos`, longitud, usuario bloqueado, tipo bloqueado).
- En segundo plano móvil: mantener expectativas realistas por restricciones del sistema.
