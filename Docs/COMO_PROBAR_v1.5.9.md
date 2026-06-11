# Cómo probar TransferLAN+ v1.5.9-beta

## Prueba simple

1. Conectar Android con Windows.
2. Verificar que se guarde la PC.
3. Cerrar Android.
4. Abrir Android.
5. Debe probar la PC conocida automáticamente.

## Prueba de IP cambiada

1. Conectar y guardar PC.
2. Cambiar la IP de la PC, por ejemplo reiniciando hotspot/router.
3. Abrir Android.
4. Debe intentar la IP anterior.
5. Si falla, debe buscar la misma PC por `device_id`.
6. Si la encuentra, actualiza la IP automáticamente.

## Resultado esperado

El usuario ve su PC, no una IP.
