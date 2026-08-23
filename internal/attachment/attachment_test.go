package attachment

import "testing"

func TestAttachmentValidation(t *testing.T) {
	item, err := BuildPhoto("V-122", "photo-1", "front.jpg", "photos/front.jpg", "image/jpeg", "photo-content")
	if err != nil {
		t.Fatal(err)
	}
	if !Verify(item, "photo-content") {
		t.Fatal("photo checksum should verify")
	}
	verified, err := MarkVerified(item, "photo-content")
	if err != nil || !verified.Verified {
		t.Fatalf("photo was not marked verified: %+v %v", verified, err)
	}
	if !ExtensionAllowed(item.Path) || !IsPhotoMedia(item.MediaType) {
		t.Fatal("photo metadata was not recognized")
	}
}

func TestAttachmentRejectsNonPhoto(t *testing.T) {
	if _, err := BuildPhoto("V-122", "photo-2", "note.txt", "docs/note.txt", "text/plain", "note"); err == nil {
		t.Fatal("non-photo attachment should be rejected")
	}
}
