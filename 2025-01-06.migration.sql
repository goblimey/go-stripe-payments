ALTER TABLE membership_sales
ADD COLUMN
IF NOT EXISTS 
ms_gift_aid boolean NOT NULL 
DEFAULT false;