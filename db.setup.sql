-- These commands create the reference data needed to run the integration tests that
-- use the database - the system user, categories, organisation, roles and item names
-- in adm_user_data.

insert into adm_users
		(usr_uuid, usr_login_name, usr_valid)
		values('System', 'System', 't');

        insert into adm_organizations(org_uuid, org_shortname, org_longname, org_homepage) 
		values('org','org','org','/');

insert into adm_categories(cat_uuid, cat_type, cat_name_intern, cat_name, 
		cat_system, cat_default, cat_sequence, cat_org_id, cat_usr_id_create)
        values('common', 'ROL', 'SYS_COMMON', 'COMMON', 
        'f', 't',1,
        (select org_id from adm_organizations limit 1),
        (select usr_id from adm_users where usr_login_name='System'));

insert into adm_categories(cat_uuid, cat_type, cat_name_intern, cat_name, 
		cat_system, cat_default, cat_sequence, cat_org_id, cat_usr_id_create)
        values('basic_data', 'USF', 'SYS_BASIC_DATA', 'BASIC_DATA', 
        'f', 't',2,
        (select org_id from adm_organizations limit 1),
        (select usr_id from adm_users where usr_login_name='System'));


insert into adm_roles(
		rol_uuid, rol_name, rol_cat_id, rol_usr_id_create, 
		rol_administrator, rol_valid) 
		values('Administrator', 'Administrator', 
        (select cat_id from adm_categories where cat_name='COMMON'),
        (select usr_id from adm_users where usr_login_name='System'), 
        't','t');

insert into adm_roles(
		rol_uuid, rol_name, rol_cat_id, rol_usr_id_create, 
		rol_administrator, rol_valid) 
		values('Member', 'Member',
        (select cat_id from adm_categories where cat_name='COMMON'), 
        (select usr_id from adm_users where usr_login_name='System'), 
        'f','t'); 

insert into adm_user_fields
(usf_uuid, usf_name, usf_name_intern, usf_type, usf_sequence, usf_cat_id, usf_usr_id_create)
values('SYS_EMAIL', 'SYS_EMAIL', 'EMAIL', 'EMAIL', 6, 
(select cat_id from adm_categories where cat_name='BASIC_DATA'),
(select usr_id from adm_users where usr_login_name='System'));

insert into adm_user_fields
(usf_uuid, usf_name, usf_name_intern, usf_type, usf_sequence, usf_cat_id, usf_usr_id_create)
values('Salutation', 'Salutation', 'SALUTATION', 'TEXT', 2,
(select cat_id from adm_categories where cat_name='BASIC_DATA'),
(select usr_id from adm_users where usr_login_name='System'));

insert into adm_user_fields
(usf_uuid, usf_name, usf_name_intern, usf_type, usf_sequence, usf_cat_id, usf_usr_id_create)
values('Initials', 'Initials', 'INITIALS', 'TEXT', 4, 
(select cat_id from adm_categories where cat_name='BASIC_DATA'), 
(select usr_id from adm_users where usr_login_name='System'));

insert into adm_user_fields
(usf_uuid, usf_name, usf_name_intern, usf_type, usf_sequence, usf_cat_id, usf_usr_id_create)
values('SYS_FIRSTNAME', 'SYS_FIRSTNAME', 'FIRST_NAME', 'TEXT', 3, 
(select cat_id from adm_categories where cat_name='BASIC_DATA'), 
(select usr_id from adm_users where usr_login_name='System'));
insert into adm_user_fields
(usf_uuid, usf_name, usf_name_intern, usf_type, usf_sequence, usf_cat_id, usf_usr_id_create)
values('SYS_LASTNAME', 'SYS_LASTNAME', 'LAST_NAME', 'TEXT', 5, 
(select cat_id from adm_categories where cat_name='BASIC_DATA'),
(select usr_id from adm_users where usr_login_name='System'));
insert into adm_user_fields
(usf_uuid, usf_name, usf_name_intern, usf_type, usf_sequence, usf_cat_id, usf_usr_id_create)
values('al1', 'Address line 1', 'STREET', 'TEXT', 7, 
(select cat_id from adm_categories where cat_name='BASIC_DATA'), 
(select usr_id from adm_users where usr_login_name='System'));


insert into adm_user_fields
(usf_uuid, usf_name, usf_name_intern, usf_type, usf_sequence, usf_cat_id, usf_usr_id_create)
values('al2', 'Address line 2', 'ADDRESS_LINE_2', 'TEXT', 8, 
(select cat_id from adm_categories where cat_name='BASIC_DATA'), 
(select usr_id from adm_users where usr_login_name='System'));
insert into adm_user_fields
(usf_uuid, usf_name, usf_name_intern, usf_type, usf_sequence, usf_cat_id, usf_usr_id_create)
values('al3', 'address line 3', 'ADDRESS_LINE_3', 'TEXT', 9, 
(select cat_id from adm_categories where cat_name='BASIC_DATA'), 
(select usr_id from adm_users where usr_login_name='System'));
insert into adm_user_fields
(usf_uuid, usf_name, usf_name_intern, usf_type, usf_sequence, usf_cat_id, usf_usr_id_create)
values('County', 'County', 'COUNTY', 'TEXT', 11, 
(select cat_id from adm_categories where cat_name='BASIC_DATA'), 
(select usr_id from adm_users where usr_login_name='System'));
insert into adm_user_fields
(usf_uuid, usf_name, usf_name_intern, usf_type, usf_sequence, usf_cat_id, usf_usr_id_create)
values('SYS_POSTCODE', 'SYS_POSTCODE', 'POSTCODE', 'TEXT', 13, 
(select cat_id from adm_categories where cat_name='BASIC_DATA'), 
(select usr_id from adm_users where usr_login_name='System'));
insert into adm_user_fields
(usf_uuid, usf_name, usf_name_intern, usf_type, usf_sequence, usf_cat_id, usf_usr_id_create)
values('SYS_COUNTRY', 'SYS_COUNTRY', 'COUNTRY', 'TEXT', 12, 
(select cat_id from adm_categories where cat_name='BASIC_DATA'), 
(select usr_id from adm_users where usr_login_name='System'));
insert into adm_user_fields
(usf_uuid, usf_name, usf_name_intern, usf_type, usf_sequence, usf_cat_id, usf_usr_id_create)
values('City', 'City', 'CITY', 'TEXT', 10, 
(select cat_id from adm_categories where cat_name='BASIC_DATA'), 
(select usr_id from adm_users where usr_login_name='System'));


insert into adm_user_fields
(usf_uuid, usf_name, usf_name_intern, usf_type, usf_sequence, usf_cat_id, usf_usr_id_create)
values('SYS_PHONE', 'SYS_PHONE', 'PHONE', 'PHONE', 15, 
(select cat_id from adm_categories where cat_name='BASIC_DATA'), 
(select usr_id from adm_users where usr_login_name='System'));
insert into adm_user_fields
(usf_uuid, usf_name, usf_name_intern, usf_type, usf_sequence, usf_cat_id, usf_usr_id_create)
values('SYS_MOBILE', 'SYS_MOBILE', 'MOBILE', 'PHONE', 16, 
(select cat_id from adm_categories where cat_name='BASIC_DATA'), 
(select usr_id from adm_users where usr_login_name='System'));
insert into adm_user_fields
(usf_uuid, usf_name, usf_name_intern, usf_type, usf_sequence, usf_cat_id, usf_usr_id_create)
values('giftaid', 'gift aid', 'GIFT_AID', 'CHECKBOX', 30, 
(select cat_id from adm_categories where cat_name='BASIC_DATA'), 
(select usr_id from adm_users where usr_login_name='System'));
insert into adm_user_fields
(usf_uuid, usf_name, usf_name_intern, usf_type, usf_sequence, usf_cat_id, usf_usr_id_create)
values('Total value of last payment', 'Total value of last payment', 'VALUE_OF_LAST_PAYMENT', 'DECIMAL', 22, 
(select cat_id from adm_categories where cat_name='BASIC_DATA'), 
(select usr_id from adm_users where usr_login_name='System'));
insert into adm_user_fields
(usf_uuid, usf_name, usf_name_intern, usf_type, usf_sequence, usf_cat_id, usf_usr_id_create)
values('fiend', 'Friend of the Museum', 'FRIEND_OF_THE_MUSEUM', 'CHECKBOX', 23, 
(select cat_id from adm_categories where cat_name='BASIC_DATA'), 
(select usr_id from adm_users where usr_login_name='System'));
insert into adm_user_fields
(usf_uuid, usf_name, usf_name_intern, usf_type, usf_sequence, usf_cat_id, usf_usr_id_create)
values('paydate', 'date last paid', 'DATE_LAST_PAID', 'DATE', 19, 
(select cat_id from adm_categories where cat_name='BASIC_DATA'), 
(select usr_id from adm_users where usr_login_name='System'));
insert into adm_user_fields
(usf_uuid, usf_name, usf_name_intern, usf_type, usf_sequence, usf_cat_id, usf_usr_id_create)
values('notices', 'Notices by email', 'NOTICES_BY_EMAIL', 'CHECKBOX', 31, 
(select cat_id from adm_categories where cat_name='BASIC_DATA'), 
(select usr_id from adm_users where usr_login_name='System'));
insert into adm_user_fields
(usf_uuid, usf_name, usf_name_intern, usf_type, usf_sequence, usf_cat_id, usf_usr_id_create)
values('members', 'Number of members of LDLHS at address', 'MEMBERS_AT_ADDRESS', 'NUMBER', 17, 
(select cat_id from adm_categories where cat_name='BASIC_DATA'), 
(select usr_id from adm_users where usr_login_name='System'));
insert into adm_user_fields
(usf_uuid, usf_name, usf_name_intern, usf_type, usf_sequence, usf_cat_id, usf_usr_id_create)
values('frends', 'Number of Friends of the Museum at this address', 'NUMBER_OF_FRIENDS_OF_THE_MUSEUM_AT_THIS_ADDRESS', 'NUMBER', 18, 
(select cat_id from adm_categories where cat_name='BASIC_DATA'), 
(select usr_id from adm_users where usr_login_name='System'));



insert into adm_user_fields
(usf_uuid, usf_name, usf_name_intern, usf_type, usf_sequence, usf_cat_id, usf_usr_id_create)
values('noticepost', 'Notices by post', 'NOTICES_BY_POST', 'CHECKBOX', 32, 
(select cat_id from adm_categories where cat_name='BASIC_DATA'), 
(select usr_id from adm_users where usr_login_name='System'));
insert into adm_user_fields
(usf_uuid, usf_name, usf_name_intern, usf_type, usf_sequence, usf_cat_id, usf_usr_id_create)
values('newsletteremail', 'Newsletter by Email', 'NEWSLETTER_BY_EMAIL', 'CHECKBOX', 33, 
(select cat_id from adm_categories where cat_name='BASIC_DATA'), 
(select usr_id from adm_users where usr_login_name='System'));
insert into adm_user_fields
(usf_uuid, usf_name, usf_name_intern, usf_type, usf_sequence, usf_cat_id, usf_usr_id_create)
values('perm', 'Permission to send emails', 'PERMISSION_TO_SEND_EMAILS', 'CHECKBOX', 33, 
(select cat_id from adm_categories where cat_name='BASIC_DATA'), 
(select usr_id from adm_users where usr_login_name='System'));
insert into adm_user_fields
(usf_uuid, usf_name, usf_name_intern, usf_type, usf_sequence, usf_cat_id, usf_usr_id_create)
values('donationSoc', 'donation to the society', 'DONATION_TO_SOCIETY', 'DECIMAL', 34, 
(select cat_id from adm_categories where cat_name='BASIC_DATA'), 
(select usr_id from adm_users where usr_login_name='System'));
insert into adm_user_fields
(usf_uuid, usf_name, usf_name_intern, usf_type, usf_sequence, usf_cat_id, usf_usr_id_create)
values('donationMus', 'Donation to the museum.', 'DONATION_TO_MUSEUM', 'DECIMAL', 35, 
(select cat_id from adm_categories where cat_name='BASIC_DATA'), 
(select usr_id from adm_users where usr_login_name='System'));

insert into adm_user_fields
(usf_uuid, usf_name, usf_name_intern, usf_type, usf_sequence, usf_cat_id, usf_usr_id_create)
values('donationSociety', 'donation to the society', 'VALUE_OF_DONATION_TO_LDLHS', 'DECIMAL', 36,
(select cat_id from adm_categories where cat_name='BASIC_DATA'), 
(select usr_id from adm_users where usr_login_name='System'));
insert into adm_user_fields
(usf_uuid, usf_name, usf_name_intern, usf_type, usf_sequence, usf_cat_id, usf_usr_id_create)
values('donationMuseum', 'Donation to the museum.', 'VALUE_OF_DONATION_TO_THE_MUSEUM', 'DECIMAL', 37, 
(select cat_id from adm_categories where cat_name='BASIC_DATA'), 
(select usr_id from adm_users where usr_login_name='System'));

insert into adm_user_fields
(usf_uuid, usf_name, usf_name_intern, usf_type, usf_sequence, usf_cat_id, usf_usr_id_create)
values('total', 'Total value of last payment', 'VALUE_OF_LAST_PAYMENT', 'DECIMAL', 38, 
(select cat_id from adm_categories where cat_name='BASIC_DATA'), 
(select usr_id from adm_users where usr_login_name='System'));
insert into adm_user_fields
(usf_uuid, usf_name, usf_name_intern, usf_type, usf_sequence, usf_cat_id, usf_usr_id_create)
values('dpperm', 'data protection permission', 'DATA_PROTECTION_PERMISSION', 'checkbox', 39, 
(select cat_id from adm_categories where cat_name='BASIC_DATA'), 
(select usr_id from adm_users where usr_login_name='System'));