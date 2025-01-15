ALTER TABLE membership_sales
ADD COLUMN
IF NOT EXISTS 
ms_giftaid boolean NOT NULL 
DEFAULT false;