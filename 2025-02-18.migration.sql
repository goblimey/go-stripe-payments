-- Add the fields to support a new member registering - name, email etc.
ALTER TABLE membership_sales
ADD COLUMN
IF NOT EXISTS 
ms_usr1_first_name varchar
(50),
ADD COLUMN
IF NOT EXISTS 
ms_usr1_last_name varchar
(50),
ADD COLUMN
IF NOT EXISTS 
ms_usr1_email varchar
(50),
ADD COLUMN
IF NOT EXISTS 
ms_usr2_first_name varchar
(50),
ADD COLUMN
IF NOT EXISTS 
ms_usr2_last_name varchar
(50),
ADD COLUMN
IF NOT EXISTS 
ms_usr2_email varchar
(50);

