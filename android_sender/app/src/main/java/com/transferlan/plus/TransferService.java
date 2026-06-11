package com.transferlan.plus;

import android.app.*;
import android.content.*;
import android.net.Uri;
import android.os.*;
import java.io.*;
import java.net.*;
import java.util.Locale;

public class TransferService extends Service {
    public static final String ACTION_START="com.transferlan.plus.START_TRANSFER";
    public static final String ACTION_CANCEL="com.transferlan.plus.CANCEL_TRANSFER";
    public static final String EXTRA_FILE_URI="file_uri";
    public static final String EXTRA_BASE_URL="base_url";
    public static final String EXTRA_FILENAME="filename";
    public static final String EXTRA_TARGET="target";
    public static final String EXTRA_SIZE="size";
    public static final String ACTION_STATUS = "com.transferlan.plus.TRANSFER_STATUS";
    public static final String EXTRA_STATUS = "status";
    public static final String EXTRA_PROGRESS = "progress";
    public static final String EXTRA_MESSAGE = "message";
    public static final String EXTRA_SENT = "sent";
    public static final String EXTRA_TOTAL = "total";
    public static final String EXTRA_RESULT = "result";
    public static final String EXTRA_SHA256 = "sha256";
    public static final String EXTRA_SERVER_FILENAME = "server_filename";
    public static final String EXTRA_SERVER_SIZE = "server_size";
    public static final String EXTRA_SERVER_PATH = "server_path";
    public static final String STATUS_PREPARING = "preparing";
    public static final String STATUS_SENDING = "sending";
    public static final String STATUS_FINALIZING = "finalizing";
    public static final String STATUS_COMPLETED = "completed";
    public static final String STATUS_ERROR = "error";
    public static final String STATUS_CANCELLED = "cancelled";

    static final String CHANNEL_ID="transferlan_transfer_service";
    static final int NOTIFICATION_ID=5051;
    static final String PREFS="transferlan_transfer_state";
    static final String KEY_STATUS="status";
    static final String KEY_PROGRESS="progress";
    static final String KEY_MESSAGE="message";
    static final String KEY_FILENAME="filename";
    static final String KEY_TARGET="target";
    static final String KEY_SENT="sent";
    static final String KEY_TOTAL="total";
    static final String KEY_TIME="time";
    static final String KEY_RESULT="result";
    static final String KEY_SHA256="sha256";
    static final String KEY_SERVER_FILENAME="server_filename";
    static final String KEY_SERVER_SIZE="server_size";
    static final String KEY_SERVER_PATH="server_path";

    NotificationManager nm;
    PowerManager.WakeLock wakeLock;
    volatile boolean cancelled=false;
    volatile HttpURLConnection conn;
    Thread worker;
    String currentFilename="";
    String currentTarget="";
    long currentTotal=0;
    String lastServerResponse="";

    public void onCreate(){
        super.onCreate();
        nm=(NotificationManager)getSystemService(Context.NOTIFICATION_SERVICE);
        if(Build.VERSION.SDK_INT>=26 && nm!=null){
            NotificationChannel ch=new NotificationChannel(CHANNEL_ID,"Transferencias en segundo plano",NotificationManager.IMPORTANCE_LOW);
            ch.setDescription("Transferencias activas de TransferLAN+");
            nm.createNotificationChannel(ch);
        }
    }

    public int onStartCommand(Intent intent,int flags,int startId){
        if(intent!=null && ACTION_CANCEL.equals(intent.getAction())){
            cancelTransfer();
            return START_NOT_STICKY;
        }
        if(intent==null || !ACTION_START.equals(intent.getAction())) return START_NOT_STICKY;
        if(worker!=null && worker.isAlive()) return START_NOT_STICKY;

        cancelled=false;
        String uriText=intent.getStringExtra(EXTRA_FILE_URI);
        String base=intent.getStringExtra(EXTRA_BASE_URL);
        String filename=intent.getStringExtra(EXTRA_FILENAME);
        String target=intent.getStringExtra(EXTRA_TARGET);
        long size=intent.getLongExtra(EXTRA_SIZE,0);

        if(filename==null||filename.length()==0) filename="archivo";
        if(target==null||target.length()==0) target="PC";

        final Uri uri=Uri.parse(uriText);
        final String fBase=trimSlash(base);
        final String fName=filename;
        final String fTarget=target;
        final long fSize=size;
        currentFilename=fName;
        currentTarget=fTarget;
        currentTotal=fSize;

        acquireWakeLock();
        startForeground(NOTIFICATION_ID, notification("Preparando transferencia",fName+" → "+fTarget,0,true));
        sendStatus(STATUS_PREPARING, 0, "Preparando " + fName, 0, fSize);

        worker=new Thread(() -> {
            long start=System.currentTimeMillis();
            try{
                upload(fBase,uri,fName,fSize,start);
                if(!cancelled) {
                    notifyNow("Transferencia completada",fName+" enviado correctamente",100,false);
                    sendStatusWithResult(STATUS_COMPLETED, 100, "Windows confirmó recepción", fSize, fSize, lastServerResponse);
                }
            }catch(Exception e){
                if(cancelled) {
                    notifyNow("Transferencia cancelada",fName,0,false);
                    sendStatus(STATUS_CANCELLED, 0, "Transferencia cancelada", 0, fSize);
                } else {
                    notifyNow("Transferencia interrumpida","Podés reintentar desde la app",0,false);
                    sendStatus(STATUS_ERROR, 0, "Transferencia interrumpida", 0, fSize);
                }
            }finally{
                releaseWakeLock();
                stopForeground(false);
                stopSelf();
            }
        });
        worker.start();
        return START_NOT_STICKY;
    }

    void cancelTransfer(){
        cancelled=true;
        try{ if(conn!=null) conn.disconnect(); }catch(Exception ignored){}
        try{ if(worker!=null) worker.interrupt(); }catch(Exception ignored){}
        notifyNow("Transferencia cancelada","El envío fue detenido.",0,false);
        sendStatus(STATUS_CANCELLED, 0, "Transferencia cancelada", 0, 0);
        releaseWakeLock();
        stopForeground(false);
        stopSelf();
    }

    void upload(String base, Uri uri, String filename, long total, long start) throws Exception{
        String boundary="TransferLANBoundary"+System.currentTimeMillis();
        conn=(HttpURLConnection)new URL(base+"/transfer/upload").openConnection();
        conn.setRequestMethod("POST");
        conn.setDoOutput(true);
        conn.setConnectTimeout(15000);
        conn.setReadTimeout(0);
        conn.setChunkedStreamingMode(1024*128);
        conn.setRequestProperty("Connection","close");
        conn.setRequestProperty("Content-Type","multipart/form-data; boundary="+boundary);

        InputStream in=getContentResolver().openInputStream(uri);
        if(in==null) throw new Exception("No se pudo abrir archivo");
        CountingOutputStream out=null;
        long last=0;
        try{
            out=new CountingOutputStream(new BufferedOutputStream(conn.getOutputStream(),1024*64));
            out.raw("--"+boundary+"\r\n");
            out.raw("Content-Disposition: form-data; name=\"file\"; filename=\""+filename+"\"\r\n");
            out.raw("Content-Type: application/octet-stream\r\n\r\n");

            byte[] buf=new byte[1024*64];
            int n;
            while(!cancelled && (n=in.read(buf))>0){
                out.writeFile(buf,0,n);
                long now=System.currentTimeMillis();
                if(now-last>500){
                    last=now;
                    int pct=total>0?(int)Math.min(99,(out.count*100)/total):0;
                    double mbps=out.count/1024.0/1024.0/Math.max(1.0,(now-start)/1000.0);
                    String msg = pct+"% · "+fmt(out.count)+" · "+String.format(Locale.US,"%.1f MB/s",mbps);
                    notifyNow("Enviando archivo",msg,pct,true);
                    sendStatus(STATUS_SENDING, pct, msg, out.count, total);
                }
            }
            if(cancelled) throw new IOException("Cancelado");
            out.raw("\r\n--"+boundary+"--\r\n");
            out.flush();
        }finally{
            try{in.close();}catch(Exception ignored){}
            try{if(out!=null)out.close();}catch(Exception ignored){}
        }

        notifyNow("Finalizando","Esperando confirmación de la PC",99,true);
        sendStatus(STATUS_FINALIZING, 99, "Esperando confirmación de la PC", total, total);
        int code=conn.getResponseCode();
        if(code<200||code>299) {
            String errBody = "";
            try { errBody = readAll(conn.getErrorStream()); } catch(Exception ignored) {}
            throw new Exception("HTTP " + code + (errBody.length() > 0 ? ": " + errBody : ""));
        }
        try{
            InputStream r=conn.getInputStream();
            lastServerResponse = readAll(r);
            r.close();
        }catch(Exception ignored){}
        conn.disconnect();
        conn=null;
    }

    Notification notification(String title,String msg,int progress,boolean ongoing){
        Intent ci=new Intent(this,TransferService.class);
        ci.setAction(ACTION_CANCEL);
        PendingIntent pi=PendingIntent.getService(this,5052,ci,Build.VERSION.SDK_INT>=23?PendingIntent.FLAG_IMMUTABLE:0);
        Notification.Builder b=Build.VERSION.SDK_INT>=26?new Notification.Builder(this,CHANNEL_ID):new Notification.Builder(this);
        b.setSmallIcon(android.R.drawable.stat_sys_upload)
         .setContentTitle("TransferLAN+ · "+title)
         .setContentText(msg)
         .setOngoing(ongoing)
         .setOnlyAlertOnce(true)
         .addAction(android.R.drawable.ic_menu_close_clear_cancel,"Cancelar",pi);
        if(ongoing) b.setProgress(100,Math.max(0,Math.min(100,progress)),false);
        return b.build();
    }


    void sendStatus(String status, int progress, String message, long sent, long total){
        try{
            getSharedPreferences(PREFS, MODE_PRIVATE)
                .edit()
                .putString(KEY_STATUS, status)
                .putInt(KEY_PROGRESS, progress)
                .putString(KEY_MESSAGE, message == null ? "" : message)
                .putString(KEY_FILENAME, currentFilename == null ? "" : currentFilename)
                .putString(KEY_TARGET, currentTarget == null ? "" : currentTarget)
                .putLong(KEY_SENT, sent)
                .putLong(KEY_TOTAL, total)
                .putLong(KEY_TIME, System.currentTimeMillis())
                .putString(KEY_RESULT, "")
                .putString(KEY_SHA256, "")
                .putString(KEY_SERVER_FILENAME, "")
                .putLong(KEY_SERVER_SIZE, 0)
                .putString(KEY_SERVER_PATH, "")
                .apply();

            Intent i = new Intent(ACTION_STATUS);
            i.setPackage(getPackageName());
            i.putExtra(EXTRA_STATUS, status);
            i.putExtra(EXTRA_PROGRESS, progress);
            i.putExtra(EXTRA_MESSAGE, message);
            i.putExtra(EXTRA_SENT, sent);
            i.putExtra(EXTRA_TOTAL, total);
            sendBroadcast(i);
        }catch(Exception ignored){}
    }

    void notifyNow(String title,String msg,int progress,boolean ongoing){
        try{ if(nm!=null) nm.notify(NOTIFICATION_ID,notification(title,msg,progress,ongoing)); }catch(Exception ignored){}
    }

    void acquireWakeLock(){
        try{
            PowerManager pm=(PowerManager)getSystemService(Context.POWER_SERVICE);
            wakeLock=pm.newWakeLock(PowerManager.PARTIAL_WAKE_LOCK,"TransferLAN+:TransferWakeLock");
            wakeLock.setReferenceCounted(false);
            wakeLock.acquire(60*60*1000L);
        }catch(Exception ignored){}
    }
    void releaseWakeLock(){try{if(wakeLock!=null&&wakeLock.isHeld())wakeLock.release();}catch(Exception ignored){}}
    public void onDestroy(){releaseWakeLock();super.onDestroy();}
    public IBinder onBind(Intent i){return null;}
    String trimSlash(String s){if(s==null)return"";while(s.endsWith("/"))s=s.substring(0,s.length()-1);return s;}
    String fmt(long b){if(b<=0)return"0 B";double v=b;String[] u={"B","KB","MB","GB","TB"};int i=0;while(v>=1024&&i<u.length-1){v/=1024;i++;}return String.format(Locale.US,"%.1f %s",v,u[i]);}

    class CountingOutputStream extends FilterOutputStream{
        long count=0;
        CountingOutputStream(OutputStream o){super(o);}
        void raw(String s)throws IOException{out.write(s.getBytes("UTF-8"));}
        void writeFile(byte[] b,int off,int len)throws IOException{out.write(b,off,len);count+=len;}
    }
}
