

insert into adm_user_fields (usf_uuid, usf_name, usf_name_intern, usf_type,
		usf_cat_id, usf_usr_id_create, usf_sequence)
values('sal', 'Salutation', 'SALUTATION', 'TEXT', 2, 1, 1);

insert into adm_user_fields (usf_uuid, usf_name, usf_name_intern, usf_type,
		usf_cat_id, usf_usr_id_create, usf_sequence)
values('init', 'Initials', 'INITIALS', 'TEXT', 4, 1, 15);

insert into adm_user_fields (usf_uuid, usf_name, usf_name_intern, usf_type,
		usf_cat_id, usf_usr_id_create, usf_sequence)
values('add2', 'Address line 2', 'ADDRESS_LINE_2', 'TEXT', 8, 1, 16);

insert into adm_user_fields (usf_uuid, usf_name, usf_name_intern, usf_type,
		usf_cat_id, usf_usr_id_create, usf_sequence)
values('add3', 'address line 3', 'ADDRESS_LINE_3', 'TEXT', 9, 1, 17);

insert into adm_user_fields (usf_uuid, usf_name, usf_name_intern, usf_type,
		usf_cat_id, usf_usr_id_create, usf_sequence)
values('cty', 'County', 'COUNTY', 'TEXT', 1, 1, 1);

insert into adm_user_fields (usf_uuid, usf_name, usf_name_intern, usf_type,
		usf_cat_id, usf_usr_id_create, usf_sequence)
values('pc', 'SYS_POSTCODE', 'POSTCODE', 'TEXT', 1, 1, 18);

insert into adm_user_fields (usf_uuid, usf_name, usf_name_intern, usf_type,
		usf_cat_id, usf_usr_id_create, usf_sequence)
values('ga', 'gift aid', 'GIFT_AID', 'CHECKBOX', 1, 1, 19);

insert into adm_user_fields (usf_uuid, usf_name, usf_name_intern, usf_type,
		usf_cat_id, usf_usr_id_create, usf_sequence)
values('fom', 'Friend of the Museum', 'FRIEND_OF_THE_MUSEUM', 'CHECKBOX', 1, 1, 20);

insert into adm_user_fields (usf_uuid, usf_name, usf_name_intern, usf_type,
		usf_cat_id, usf_usr_id_create, usf_sequence)
values('dlp', 'date last paid', 'DATE_LAST_PAID', 'DATE', 1, 1, 21);

insert into adm_user_fields (usf_uuid, usf_name, usf_name_intern, usf_type,
		usf_cat_id, usf_usr_id_create, usf_sequence)
values('ne', 'Notices by email', 'NOTICES_BY_EMAIL', 'CHECKBOX', 1, 1, 22);

insert into adm_user_fields (usf_uuid, usf_name, usf_name_intern, usf_type,
		usf_cat_id, usf_usr_id_create, usf_sequence)
values('nummem', 'Number of members of LDLHS at address', 'MEMBERS_AT_ADDRESS', 'NUMBER', 1, 1, 23);

insert into adm_user_fields (usf_uuid, usf_name, usf_name_intern, usf_type,
		usf_cat_id, usf_usr_id_create, usf_sequence)
values('numfr', 'Number of Friends of the Museum at this address', 
'NUMBER_OF_FRIENDS_OF_THE_MUSEUM_AT_THIS_ADDRESS', 'NUMBER', 1, 1, 24);

insert into adm_user_fields (usf_uuid, usf_name, usf_name_intern, usf_type,
		usf_cat_id, usf_usr_id_create, usf_sequence)
values('nbe', 'Newsletter by Email', 'NEWSLETTER_BY_EMAIL', 'CHECKBOX', 1, 1, 25);

insert into adm_user_fields (usf_uuid, usf_name, usf_name_intern, usf_type,
		usf_cat_id, usf_usr_id_create, usf_sequence)
values('pse', 'Permission to send emails', 'PERMISSION_TO_SEND_EMAILS', 'CHECKBOX', 1, 1, 26);

insert into adm_user_fields (usf_uuid, usf_name, usf_name_intern, usf_type,
		usf_cat_id, usf_usr_id_create, usf_sequence)
values('d2s', 'donation to the society', 'VALUE_OF_DONATION_TO_LDLHS', 'DECIMAL', 1, 1, 27);

insert into adm_user_fields (usf_uuid, usf_name, usf_name_intern, usf_type,
		usf_cat_id, usf_usr_id_create, usf_sequence)
values('d2m', 'Donation to the museum.', 'VALUE_OF_DONATION_TO_THE_MUSEUM', 'DECIMAL', 1, 1, 28);

insert into adm_user_fields (usf_uuid, usf_name, usf_name_intern, usf_type,
		usf_cat_id, usf_usr_id_create, usf_sequence)
values('tvlp', 'Total value of last payment', 'VALUE_OF_LAST_PAYMENT', 'DECIMAL', 1, 1, 29);

insert into adm_user_fields (usf_uuid, usf_name, usf_name_intern, usf_type,
		usf_cat_id, usf_usr_id_create, usf_sequence)
values('loi', 'Location of Interest', 'LOCATION_OF_INTEREST', 'checkbox', 1, 1, 30);

		