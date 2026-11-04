package file

import (
	"errors"
	"testing"

	"TDrive/backend/projection"
)

func TestForkArchiveLimitsAcceptLargeEncryptedFiles(t *testing.T) {
	service := &Service{}
	if LargeFileMaxBytes != 256<<30 {
		t.Fatalf("fork hard cap = %d, want 256 GiB", LargeFileMaxBytes)
	}
	if MaxParts < 1024 {
		t.Fatalf("fork part count = %d, want at least 1024", MaxParts)
	}

	// Size calculations and part plans only; these tests do not allocate
	// or transmit huge video files.
	for _, size := range []int64{41 << 30, 73 << 30, 128 << 30, 255 << 30} {
		stored, multipart, err := service.planUpload("archive.mp4", size, true)
		if err != nil || !multipart || stored < size {
			t.Fatalf("%d-byte encrypted file: stored=%d multi=%t err=%v", size, stored, multipart, err)
		}
		plan, err := service.buildResumablePartPlan(stored)
		if err != nil || plan.partCount < 2 || plan.partCount > MaxParts ||
			plan.partSize > MaxPartBytes {
			t.Fatalf("resumable %d bytes: plan=%+v err=%v", size, plan, err)
		}
		last, length, err := plan.window(stored, plan.partCount-1)
		if err != nil || last+length != stored {
			t.Fatalf("last part doesn't span %d bytes", stored)
		}
	}
	for _, encrypted := range []bool{false, true} {
		if _, _, err := service.planUpload("too-big.mp4", LargeFileMaxBytes+1, encrypted); !errors.Is(err, ErrFileTooLarge) {
			t.Fatalf("oversize encrypted=%v: got %v, want ErrFileTooLarge", encrypted, err)
		}
	}
}

func TestForkArchiveLargeDownloadPlan(t *testing.T) {
	const total int64 = 73 << 30
	const partSize int64 = 1 << 30
	parts := make([]projection.FilePart, 0, 73)
	for idx := 0; idx < 73; idx++ {
		parts = append(parts, projection.FilePart{
			PartIndex: idx, MsgID: int64(idx + 10), Size: partSize,
		})
	}
	manifest := projection.DownloadFile{StoredSize: total, Parts: parts}
	blocks, ids, err := planDownloadBlocks(manifest, 1<<20)
	if err != nil {
		t.Fatalf("download plan: %v", err)
	}
	if len(ids) != 73 || len(blocks) != 73*1024 {
		t.Fatalf("unexpected planned blocks: ids=%d blocks=%d", len(ids), len(blocks))
	}
	manifest.StoredSize = LargeFileMaxBytes + 1
	if _, _, err := planDownloadBlocks(manifest, 1<<20); err == nil {
		t.Fatal("overlimit download should fail")
	}
}
