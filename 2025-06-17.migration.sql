-- Change these columns to the form used for similar columns.
ALTER TABLE membership_sales 
RENAME COLUMN ms_user1_email 
TO ms_usr1_email;
ALTER TABLE membership_sales 
RENAME COLUMN timestamp_create 
TO ms_timestamp_create;

