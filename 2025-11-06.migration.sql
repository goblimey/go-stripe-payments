-- Unique constraint - only one interest per user and interest.
ALTER TABLE public.adm_members_interests
ADD CONSTRAINT adm_un_usr_interest UNIQUE (mi_usr_id, mi_interest_id);