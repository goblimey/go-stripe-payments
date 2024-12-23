alter table only public.membership_sales
drop constraint adm_fk_ms_usr2_id;
alter table only public.membership_sales
drop constraint adm_fk_ms_usr1_id;
alter table only public.membership_sales
drop constraint membership_sales_pkey;
drop sequence public.membership_sales_ms_id_seq
cascade;
drop table membership_sales;