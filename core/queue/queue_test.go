package queue

import "testing"

func TestQueueAddUpdateCancelClear(t *testing.T) {
    q := NewManager()
    item := q.Add("video.mp4", "http://pc:47231", 100)
    if item.Status != StatusPending { t.Fatalf("expected pending, got %s", item.Status) }
    if len(q.List()) != 1 { t.Fatalf("expected 1 item") }
    if err := q.UpdateProgress(item.ID, 50, StatusRunning, ""); err != nil { t.Fatal(err) }
    if q.List()[0].SentBytes != 50 { t.Fatalf("progress was not updated") }
    if err := q.Cancel(item.ID); err != nil { t.Fatal(err) }
    if q.List()[0].Status != StatusCanceled { t.Fatalf("expected canceled") }
    if removed := q.ClearFinished(); removed != 1 { t.Fatalf("expected 1 removed, got %d", removed) }
}
