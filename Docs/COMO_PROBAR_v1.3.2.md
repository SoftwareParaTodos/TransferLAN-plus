# Cómo probar TransferLAN+ v1.3.2-beta

## Windows

1. Ejecutar:

```text
Windows/INICIAR_TRANSFERLAN.bat
```

2. Abrir:

```text
http://localhost:5050
```

3. Verificar:
- estado activo;
- QR visible;
- código de PC;
- historial.

## Android

1. Instalar APK generado por GitHub Actions.
2. Abrir TransferLAN+.
3. Conectar usando:
- Reconectar última PC;
- Agregar PC por IP;
- Pegar código de PC.

4. Elegir archivo.
5. Enviar.
6. Confirmar:
- progreso;
- mensaje final;
- historial local;
- archivo recibido en Windows.

## Si no conecta

Ejecutar como administrador:

```text
Windows/PERMITIR_FIREWALL_ADMIN.bat
```

Y probar desde Chrome del celular:

```text
http://IP-DE-LA-PC:5050/device/info
```
