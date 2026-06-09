# TransferLAN+ v0.13.0 - Drag & Drop + Historial

## Objetivo

Convertir la interfaz en una experiencia más cómoda de escritorio:

- Arrastrar un archivo sobre la ventana para enviarlo.
- Mantener un historial local de transferencias.
- Consultar y limpiar el historial por API.
- Registrar recepciones normales, por bloques y por Modo Invitado Web.

## Endpoints nuevos

### GET /history

Devuelve las últimas transferencias registradas.

Parámetros opcionales:

```text
limit=50
```

Ejemplo:

```bash
curl http://127.0.0.1:47231/history
```

### POST /history/clear

Limpia el historial local.

```bash
curl -X POST http://127.0.0.1:47231/history/clear
```

## Archivo local

El historial se guarda en:

```text
data/history.json
```

Por ahora es JSON para que sea fácil de revisar y depurar. Más adelante puede migrarse a SQLite cuando haga falta indexar, filtrar o sincronizar más datos.

## Drag & Drop

La UI Flutter incorpora una zona de drop en la pantalla principal.

Flujo esperado:

1. Abrir TransferLAN+ desktop.
2. Buscar dispositivos LAN o configurar receptor manual.
3. Arrastrar un archivo sobre la ventana.
4. Se abre la pantalla de envío.
5. Confirmar envío.

## Estado

Esta versión prioriza comodidad de uso. No reemplaza todavía el instalador final ni las notificaciones nativas.
