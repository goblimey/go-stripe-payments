-- Add ms_transaction_type and set all existing records to
-- 'membership renewal'.
ALTER TABLE membership_sales
ADD COLUMN IF NOT EXISTS 
ms_transaction_type varchar (30) NOT NULL
DEFAULT 'membership renewal';

