package com.transferlan.plus;

import android.app.Notification;
import android.app.NotificationChannel;
import android.app.NotificationManager;
import android.app.PendingIntent;
import android.app.Service;
import android.content.Intent;
import android.os.Build;
import android.os.IBinder;
import android.os.PowerManager;
import android.content.Context;

public class TransferService extends Service {
    public static final String ACTION_START = "com.transferlan.plus.START_TRANSFER";
    public static final String ACTION_CANCEL = "com.transferlan.plus.CANCEL_TRANSFER";
    public static final String EXTRA_FILENAME = "filename";
    public static final String EXTRA_TARGET = "target";

    private static final String CHANNEL_ID = "transferlan_transfer_service";
    private static final int NOTIFICATION_ID = 5051;

    private NotificationManager notificationManager;
    private PowerManager.WakeLock wakeLock;
    private boolean cancelled = false;

    @Override
    public void onCreate() {
        super.onCreate();
        notificationManager = (NotificationManager) getSystemService(Context.NOTIFICATION_SERVICE);
        createChannel();
    }

    @Override
    public int onStartCommand(Intent intent, int flags, int startId) {
        if (intent != null && ACTION_CANCEL.equals(intent.getAction())) {
            cancelled = true;
            updateNotification("Transferencia cancelada", "El envío fue detenido.", 0, false);
            stopForeground(false);
            stopSelf();
            return START_NOT_STICKY;
        }

        String filename = intent != null ? intent.getStringExtra(EXTRA_FILENAME) : null;
        String target = intent != null ? intent.getStringExtra(EXTRA_TARGET) : null;

        if (filename == null || filename.length() == 0) filename = "archivo";
        if (target == null || target.length() == 0) target = "PC";

        acquireWakeLock();
        startForeground(NOTIFICATION_ID, buildNotification("Preparando transferencia", filename + " → " + target, 0, true));

        /*
         * v1.5.1-beta:
         * Este servicio deja preparada la arquitectura para mover el upload pesado
         * fuera de MainActivity.
         *
         * En v1.5.2-beta se migrará el envío completo a este servicio:
         * - URI del archivo
         * - destino base_url
         * - streaming HTTP
         * - progreso por notificación
         * - cancelación real
         * - recuperación de estado al reabrir la app
         */

        return START_NOT_STICKY;
    }

    @Override
    public void onDestroy() {
        releaseWakeLock();
        super.onDestroy();
    }

    @Override
    public IBinder onBind(Intent intent) {
        return null;
    }

    private void createChannel() {
        if (Build.VERSION.SDK_INT >= 26 && notificationManager != null) {
            NotificationChannel channel = new NotificationChannel(
                CHANNEL_ID,
                "Transferencias en segundo plano",
                NotificationManager.IMPORTANCE_LOW
            );
            channel.setDescription("Transferencias activas de TransferLAN+");
            notificationManager.createNotificationChannel(channel);
        }
    }

    private Notification buildNotification(String title, String message, int progress, boolean ongoing) {
        Intent cancelIntent = new Intent(this, TransferService.class);
        cancelIntent.setAction(ACTION_CANCEL);

        PendingIntent cancelPendingIntent = PendingIntent.getService(
            this,
            5052,
            cancelIntent,
            Build.VERSION.SDK_INT >= 23 ? PendingIntent.FLAG_IMMUTABLE : 0
        );

        Notification.Builder builder = Build.VERSION.SDK_INT >= 26
            ? new Notification.Builder(this, CHANNEL_ID)
            : new Notification.Builder(this);

        builder.setSmallIcon(android.R.drawable.stat_sys_upload)
            .setContentTitle("TransferLAN+ · " + title)
            .setContentText(message)
            .setOngoing(ongoing)
            .setOnlyAlertOnce(true)
            .addAction(android.R.drawable.ic_menu_close_clear_cancel, "Cancelar", cancelPendingIntent);

        if (ongoing) {
            builder.setProgress(100, Math.max(0, Math.min(100, progress)), false);
        }

        return builder.build();
    }

    public void updateNotification(String title, String message, int progress, boolean ongoing) {
        try {
            if (notificationManager != null) {
                notificationManager.notify(NOTIFICATION_ID, buildNotification(title, message, progress, ongoing));
            }
        } catch(Exception ignored) {}
    }

    private void acquireWakeLock() {
        try {
            PowerManager pm = (PowerManager) getSystemService(Context.POWER_SERVICE);
            if (pm != null) {
                wakeLock = pm.newWakeLock(PowerManager.PARTIAL_WAKE_LOCK, "TransferLAN+:TransferWakeLock");
                wakeLock.setReferenceCounted(false);
                wakeLock.acquire(60 * 60 * 1000L);
            }
        } catch(Exception ignored) {}
    }

    private void releaseWakeLock() {
        try {
            if (wakeLock != null && wakeLock.isHeld()) {
                wakeLock.release();
            }
        } catch(Exception ignored) {}
    }
}
