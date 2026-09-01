-- +goose Up
CREATE TABLE voc_status (
    code TEXT PRIMARY KEY,
    label TEXT NOT NULL,
    is_closed BOOLEAN NOT NULL DEFAULT false
);

INSERT INTO voc_status (code, label, is_closed) VALUES
    ('SUBMITTED', 'รับเรื่องแล้ว', false),
    ('ACKNOWLEDGED', 'รับทราบแล้ว', false),
    ('IN_PROGRESS', 'อยู่ระหว่างดำเนินการ', false),
    ('WAITING_CUSTOMER', 'รอข้อมูลเพิ่มเติมจากผู้แจ้ง', false),
    ('RESOLVED', 'ดำเนินการเสร็จสิ้น', true),
    ('REJECTED', 'ปฏิเสธเรื่อง', true),
    ('CANCELLED', 'ยกเลิกเรื่อง', true);

-- +goose Down
DROP TABLE voc_status;
