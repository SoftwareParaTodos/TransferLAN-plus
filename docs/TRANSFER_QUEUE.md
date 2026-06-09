# TransferLAN+ v0.15.0 - Cola de transferencias

Esta versión agrega una cola local para administrar varias transferencias a la vez.

## Objetivo

Evitar que la aplicación se vuelva inestable cuando el usuario envía varios archivos grandes, carpetas o transferencias invitadas en paralelo.

## Estados

- `pending`: pendiente
- `running`: en progreso
- `completed`: finalizada
- `failed`: fallida
- `canceled`: cancelada

## Endpoints

- `GET /queue`: lista la cola.
- `POST /queue/add`: agrega una transferencia.
- `POST /queue/progress`: actualiza progreso/estado.
- `POST /queue/cancel`: cancela un ítem pendiente o en progreso.
- `POST /queue/clear-finished`: limpia finalizadas, fallidas y canceladas.

## Próximo paso

Conectar el envío por bloques real con la cola para limitar la cantidad de transferencias simultáneas y mostrar progreso en la UI Flutter.
