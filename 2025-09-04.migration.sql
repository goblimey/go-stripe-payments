-- Remove the NOT NULL constraint.  When a new user opens an account, the 
-- membershipsale record is created before the users.
alter table membership_sales alter column ms_usr1_id DROP NOT NULL;
alter table membership_sales alter column ms_usr1_id SET DEFAULT NULL;
