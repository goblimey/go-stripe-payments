--
-- PostgreSQL database dump
--

-- Dumped from database version 17.4 EE 1.4.1 (Debian 17.4ee1.4.1-1.pgee12~demo+1)
-- Dumped by pg_dump version 17.4 EE 1.4.1 (Debian 17.4ee1.4.1-1.pgee12~demo+1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: adm_announcements; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.adm_announcements (
    ann_id integer NOT NULL,
    ann_cat_id integer NOT NULL,
    ann_uuid character varying(36) NOT NULL,
    ann_headline character varying(100) NOT NULL,
    ann_description text,
    ann_usr_id_create integer,
    ann_timestamp_create timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    ann_usr_id_change integer,
    ann_timestamp_change timestamp without time zone
);


ALTER TABLE public.adm_announcements OWNER TO postgres;

--
-- Name: adm_announcements_ann_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.adm_announcements_ann_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.adm_announcements_ann_id_seq OWNER TO postgres;

--
-- Name: adm_announcements_ann_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.adm_announcements_ann_id_seq OWNED BY public.adm_announcements.ann_id;


--
-- Name: adm_auto_login; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.adm_auto_login (
    atl_id integer NOT NULL,
    atl_auto_login_id character varying(255) NOT NULL,
    atl_session_id character varying(255) NOT NULL,
    atl_org_id integer NOT NULL,
    atl_usr_id integer NOT NULL,
    atl_last_login timestamp without time zone,
    atl_number_invalid smallint DEFAULT 0 NOT NULL
);


ALTER TABLE public.adm_auto_login OWNER TO postgres;

--
-- Name: adm_auto_login_atl_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.adm_auto_login_atl_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.adm_auto_login_atl_id_seq OWNER TO postgres;

--
-- Name: adm_auto_login_atl_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.adm_auto_login_atl_id_seq OWNED BY public.adm_auto_login.atl_id;


--
-- Name: adm_categories; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.adm_categories (
    cat_id integer NOT NULL,
    cat_org_id integer,
    cat_uuid character varying(36) NOT NULL,
    cat_type character varying(10) NOT NULL,
    cat_name_intern character varying(110) NOT NULL,
    cat_name character varying(100) NOT NULL,
    cat_system boolean DEFAULT false NOT NULL,
    cat_default boolean DEFAULT false NOT NULL,
    cat_sequence smallint NOT NULL,
    cat_usr_id_create integer,
    cat_timestamp_create timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    cat_usr_id_change integer,
    cat_timestamp_change timestamp without time zone
);


ALTER TABLE public.adm_categories OWNER TO postgres;

--
-- Name: adm_categories_cat_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.adm_categories_cat_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.adm_categories_cat_id_seq OWNER TO postgres;

--
-- Name: adm_categories_cat_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.adm_categories_cat_id_seq OWNED BY public.adm_categories.cat_id;


--
-- Name: adm_category_report; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.adm_category_report (
    crt_id integer NOT NULL,
    crt_org_id integer,
    crt_name character varying(100) NOT NULL,
    crt_col_fields text,
    crt_selection_role character varying(100),
    crt_selection_cat character varying(100),
    crt_number_col boolean DEFAULT false NOT NULL
);


ALTER TABLE public.adm_category_report OWNER TO postgres;

--
-- Name: adm_category_report_crt_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.adm_category_report_crt_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.adm_category_report_crt_id_seq OWNER TO postgres;

--
-- Name: adm_category_report_crt_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.adm_category_report_crt_id_seq OWNED BY public.adm_category_report.crt_id;


--
-- Name: adm_components; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.adm_components (
    com_id integer NOT NULL,
    com_type character varying(10) NOT NULL,
    com_name character varying(255) NOT NULL,
    com_name_intern character varying(255) NOT NULL,
    com_version character varying(10) NOT NULL,
    com_beta smallint DEFAULT 0 NOT NULL,
    com_update_step integer DEFAULT 0 NOT NULL,
    com_update_completed boolean DEFAULT true NOT NULL,
    com_timestamp_installed timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.adm_components OWNER TO postgres;

--
-- Name: adm_components_com_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.adm_components_com_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.adm_components_com_id_seq OWNER TO postgres;

--
-- Name: adm_components_com_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.adm_components_com_id_seq OWNED BY public.adm_components.com_id;


--
-- Name: adm_events; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.adm_events (
    dat_id integer NOT NULL,
    dat_cat_id integer NOT NULL,
    dat_rol_id integer,
    dat_room_id integer,
    dat_uuid character varying(36) NOT NULL,
    dat_begin timestamp without time zone,
    dat_end timestamp without time zone,
    dat_all_day boolean DEFAULT false NOT NULL,
    dat_headline character varying(100) NOT NULL,
    dat_description text,
    dat_highlight boolean DEFAULT false NOT NULL,
    dat_location character varying(100),
    dat_country character varying(100),
    dat_deadline timestamp without time zone,
    dat_max_members integer DEFAULT 0 NOT NULL,
    dat_usr_id_create integer,
    dat_timestamp_create timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    dat_usr_id_change integer,
    dat_timestamp_change timestamp without time zone,
    dat_allow_comments boolean DEFAULT false NOT NULL,
    dat_additional_guests boolean DEFAULT false NOT NULL
);


ALTER TABLE public.adm_events OWNER TO postgres;

--
-- Name: adm_events_dat_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.adm_events_dat_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.adm_events_dat_id_seq OWNER TO postgres;

--
-- Name: adm_events_dat_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.adm_events_dat_id_seq OWNED BY public.adm_events.dat_id;


--
-- Name: adm_files; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.adm_files (
    fil_id integer NOT NULL,
    fil_fol_id integer NOT NULL,
    fil_uuid character varying(36) NOT NULL,
    fil_name character varying(255) NOT NULL,
    fil_description text,
    fil_locked boolean DEFAULT false NOT NULL,
    fil_counter integer,
    fil_usr_id integer,
    fil_timestamp timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.adm_files OWNER TO postgres;

--
-- Name: adm_files_fil_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.adm_files_fil_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.adm_files_fil_id_seq OWNER TO postgres;

--
-- Name: adm_files_fil_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.adm_files_fil_id_seq OWNED BY public.adm_files.fil_id;


--
-- Name: adm_folders; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.adm_folders (
    fol_id integer NOT NULL,
    fol_org_id integer NOT NULL,
    fol_fol_id_parent integer,
    fol_uuid character varying(36) NOT NULL,
    fol_type character varying(10) NOT NULL,
    fol_name character varying(255) NOT NULL,
    fol_description text,
    fol_path character varying(255) NOT NULL,
    fol_locked boolean DEFAULT false NOT NULL,
    fol_public boolean DEFAULT false NOT NULL,
    fol_usr_id integer,
    fol_timestamp timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.adm_folders OWNER TO postgres;

--
-- Name: adm_folders_fol_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.adm_folders_fol_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.adm_folders_fol_id_seq OWNER TO postgres;

--
-- Name: adm_folders_fol_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.adm_folders_fol_id_seq OWNED BY public.adm_folders.fol_id;


--
-- Name: adm_guestbook; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.adm_guestbook (
    gbo_id integer NOT NULL,
    gbo_org_id integer NOT NULL,
    gbo_uuid character varying(36) NOT NULL,
    gbo_name character varying(60) NOT NULL,
    gbo_text text NOT NULL,
    gbo_email character varying(254),
    gbo_homepage character varying(50),
    gbo_ip_address character varying(39) NOT NULL,
    gbo_locked boolean DEFAULT false NOT NULL,
    gbo_usr_id_create integer,
    gbo_timestamp_create timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    gbo_usr_id_change integer,
    gbo_timestamp_change timestamp without time zone
);


ALTER TABLE public.adm_guestbook OWNER TO postgres;

--
-- Name: adm_guestbook_comments; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.adm_guestbook_comments (
    gbc_id integer NOT NULL,
    gbc_gbo_id integer NOT NULL,
    gbc_uuid character varying(36) NOT NULL,
    gbc_name character varying(60) NOT NULL,
    gbc_text text NOT NULL,
    gbc_email character varying(254),
    gbc_ip_address character varying(39) NOT NULL,
    gbc_locked boolean DEFAULT false NOT NULL,
    gbc_usr_id_create integer,
    gbc_timestamp_create timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    gbc_usr_id_change integer,
    gbc_timestamp_change timestamp without time zone
);


ALTER TABLE public.adm_guestbook_comments OWNER TO postgres;

--
-- Name: adm_guestbook_comments_gbc_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.adm_guestbook_comments_gbc_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.adm_guestbook_comments_gbc_id_seq OWNER TO postgres;

--
-- Name: adm_guestbook_comments_gbc_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.adm_guestbook_comments_gbc_id_seq OWNED BY public.adm_guestbook_comments.gbc_id;


--
-- Name: adm_guestbook_gbo_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.adm_guestbook_gbo_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.adm_guestbook_gbo_id_seq OWNER TO postgres;

--
-- Name: adm_guestbook_gbo_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.adm_guestbook_gbo_id_seq OWNED BY public.adm_guestbook.gbo_id;


--
-- Name: adm_ids; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.adm_ids (
    ids_usr_id integer NOT NULL,
    ids_reference_id integer NOT NULL
);


ALTER TABLE public.adm_ids OWNER TO postgres;

--
-- Name: adm_interests_ntrst_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.adm_interests_ntrst_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.adm_interests_ntrst_id_seq OWNER TO postgres;

--
-- Name: adm_interests; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.adm_interests (
    ntrst_id integer DEFAULT nextval('public.adm_interests_ntrst_id_seq'::regclass) NOT NULL,
    ntrst_name character varying(50)
);


ALTER TABLE public.adm_interests OWNER TO postgres;

--
-- Name: adm_links; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.adm_links (
    lnk_id integer NOT NULL,
    lnk_cat_id integer NOT NULL,
    lnk_uuid character varying(36) NOT NULL,
    lnk_name character varying(255) NOT NULL,
    lnk_description text,
    lnk_url character varying(2000) NOT NULL,
    lnk_counter integer DEFAULT 0 NOT NULL,
    lnk_usr_id_create integer,
    lnk_timestamp_create timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    lnk_usr_id_change integer,
    lnk_timestamp_change timestamp without time zone
);


ALTER TABLE public.adm_links OWNER TO postgres;

--
-- Name: adm_links_lnk_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.adm_links_lnk_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.adm_links_lnk_id_seq OWNER TO postgres;

--
-- Name: adm_links_lnk_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.adm_links_lnk_id_seq OWNED BY public.adm_links.lnk_id;


--
-- Name: adm_list_columns; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.adm_list_columns (
    lsc_id integer NOT NULL,
    lsc_lst_id integer NOT NULL,
    lsc_number smallint NOT NULL,
    lsc_usf_id integer,
    lsc_special_field character varying(255),
    lsc_sort character varying(5),
    lsc_filter character varying(255)
);


ALTER TABLE public.adm_list_columns OWNER TO postgres;

--
-- Name: adm_list_columns_lsc_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.adm_list_columns_lsc_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.adm_list_columns_lsc_id_seq OWNER TO postgres;

--
-- Name: adm_list_columns_lsc_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.adm_list_columns_lsc_id_seq OWNED BY public.adm_list_columns.lsc_id;


--
-- Name: adm_lists; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.adm_lists (
    lst_id integer NOT NULL,
    lst_org_id integer NOT NULL,
    lst_usr_id integer NOT NULL,
    lst_uuid character varying(36) NOT NULL,
    lst_name character varying(255),
    lst_timestamp timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    lst_global boolean DEFAULT false NOT NULL
);


ALTER TABLE public.adm_lists OWNER TO postgres;

--
-- Name: adm_lists_lst_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.adm_lists_lst_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.adm_lists_lst_id_seq OWNER TO postgres;

--
-- Name: adm_lists_lst_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.adm_lists_lst_id_seq OWNED BY public.adm_lists.lst_id;


--
-- Name: adm_members; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.adm_members (
    mem_id integer NOT NULL,
    mem_rol_id integer NOT NULL,
    mem_usr_id integer NOT NULL,
    mem_uuid character varying(36) NOT NULL,
    mem_begin date NOT NULL,
    mem_end date DEFAULT '9999-12-31'::date NOT NULL,
    mem_leader boolean DEFAULT false NOT NULL,
    mem_usr_id_create integer,
    mem_timestamp_create timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    mem_usr_id_change integer,
    mem_timestamp_change timestamp without time zone,
    mem_approved integer,
    mem_comment character varying(4000),
    mem_count_guests integer DEFAULT 0 NOT NULL
);


ALTER TABLE public.adm_members OWNER TO postgres;

--
-- Name: adm_members_interests; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.adm_members_interests (
    mi_id integer NOT NULL,
    mi_usr_id integer NOT NULL,
    mi_interest_id integer NOT NULL
);


ALTER TABLE public.adm_members_interests OWNER TO postgres;

--
-- Name: adm_members_interests_mi_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.adm_members_interests_mi_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.adm_members_interests_mi_id_seq OWNER TO postgres;

--
-- Name: adm_members_interests_mi_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.adm_members_interests_mi_id_seq OWNED BY public.adm_members_interests.mi_id;


--
-- Name: adm_members_mem_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.adm_members_mem_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.adm_members_mem_id_seq OWNER TO postgres;

--
-- Name: adm_members_mem_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.adm_members_mem_id_seq OWNED BY public.adm_members.mem_id;


--
-- Name: adm_members_other_interests; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.adm_members_other_interests (
    moi_id integer NOT NULL,
    moi_usr_id integer NOT NULL,
    moi_interests character varying(200)
);


ALTER TABLE public.adm_members_other_interests OWNER TO postgres;

--
-- Name: adm_members_other_interests_moi_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.adm_members_other_interests_moi_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.adm_members_other_interests_moi_id_seq OWNER TO postgres;

--
-- Name: adm_members_other_interests_moi_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.adm_members_other_interests_moi_id_seq OWNED BY public.adm_members_other_interests.moi_id;


--
-- Name: adm_menu; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.adm_menu (
    men_id integer NOT NULL,
    men_men_id_parent integer,
    men_com_id integer,
    men_uuid character varying(36) NOT NULL,
    men_name_intern character varying(255),
    men_name character varying(255),
    men_description character varying(4000),
    men_node boolean DEFAULT false NOT NULL,
    men_order integer,
    men_standard boolean DEFAULT false NOT NULL,
    men_url character varying(2000),
    men_icon character varying(100)
);


ALTER TABLE public.adm_menu OWNER TO postgres;

--
-- Name: adm_menu_men_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.adm_menu_men_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.adm_menu_men_id_seq OWNER TO postgres;

--
-- Name: adm_menu_men_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.adm_menu_men_id_seq OWNED BY public.adm_menu.men_id;


--
-- Name: adm_messages; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.adm_messages (
    msg_id integer NOT NULL,
    msg_uuid character varying(36) NOT NULL,
    msg_type character varying(10) NOT NULL,
    msg_subject character varying(256) NOT NULL,
    msg_usr_id_sender integer NOT NULL,
    msg_timestamp timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    msg_read smallint DEFAULT 0 NOT NULL
);


ALTER TABLE public.adm_messages OWNER TO postgres;

--
-- Name: adm_messages_attachments; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.adm_messages_attachments (
    msa_id integer NOT NULL,
    msa_msg_id integer NOT NULL,
    msa_file_name character varying(256) NOT NULL,
    msa_original_file_name character varying(256) NOT NULL
);


ALTER TABLE public.adm_messages_attachments OWNER TO postgres;

--
-- Name: adm_messages_attachments_msa_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.adm_messages_attachments_msa_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.adm_messages_attachments_msa_id_seq OWNER TO postgres;

--
-- Name: adm_messages_attachments_msa_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.adm_messages_attachments_msa_id_seq OWNED BY public.adm_messages_attachments.msa_id;


--
-- Name: adm_messages_content; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.adm_messages_content (
    msc_id integer NOT NULL,
    msc_msg_id integer NOT NULL,
    msc_usr_id integer,
    msc_message text NOT NULL,
    msc_timestamp timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.adm_messages_content OWNER TO postgres;

--
-- Name: adm_messages_content_msc_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.adm_messages_content_msc_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.adm_messages_content_msc_id_seq OWNER TO postgres;

--
-- Name: adm_messages_content_msc_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.adm_messages_content_msc_id_seq OWNED BY public.adm_messages_content.msc_id;


--
-- Name: adm_messages_msg_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.adm_messages_msg_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.adm_messages_msg_id_seq OWNER TO postgres;

--
-- Name: adm_messages_msg_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.adm_messages_msg_id_seq OWNED BY public.adm_messages.msg_id;


--
-- Name: adm_messages_recipients; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.adm_messages_recipients (
    msr_id integer NOT NULL,
    msr_msg_id integer NOT NULL,
    msr_rol_id integer,
    msr_usr_id integer,
    msr_role_mode smallint DEFAULT 0 NOT NULL
);


ALTER TABLE public.adm_messages_recipients OWNER TO postgres;

--
-- Name: adm_messages_recipients_msr_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.adm_messages_recipients_msr_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.adm_messages_recipients_msr_id_seq OWNER TO postgres;

--
-- Name: adm_messages_recipients_msr_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.adm_messages_recipients_msr_id_seq OWNED BY public.adm_messages_recipients.msr_id;


--
-- Name: adm_organizations; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.adm_organizations (
    org_id integer NOT NULL,
    org_uuid character varying(36) NOT NULL,
    org_shortname character varying(10) NOT NULL,
    org_longname character varying(60) NOT NULL,
    org_org_id_parent integer,
    org_homepage character varying(60) NOT NULL
);


ALTER TABLE public.adm_organizations OWNER TO postgres;

--
-- Name: adm_organizations_org_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.adm_organizations_org_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.adm_organizations_org_id_seq OWNER TO postgres;

--
-- Name: adm_organizations_org_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.adm_organizations_org_id_seq OWNED BY public.adm_organizations.org_id;


--
-- Name: adm_photos; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.adm_photos (
    pho_id integer NOT NULL,
    pho_org_id integer NOT NULL,
    pho_pho_id_parent integer,
    pho_uuid character varying(36) NOT NULL,
    pho_quantity integer DEFAULT 0 NOT NULL,
    pho_name character varying(50) NOT NULL,
    pho_begin date NOT NULL,
    pho_end date NOT NULL,
    pho_description character varying(4000),
    pho_photographers character varying(100),
    pho_locked boolean DEFAULT false NOT NULL,
    pho_usr_id_create integer,
    pho_timestamp_create timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    pho_usr_id_change integer,
    pho_timestamp_change timestamp without time zone
);


ALTER TABLE public.adm_photos OWNER TO postgres;

--
-- Name: adm_photos_pho_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.adm_photos_pho_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.adm_photos_pho_id_seq OWNER TO postgres;

--
-- Name: adm_photos_pho_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.adm_photos_pho_id_seq OWNED BY public.adm_photos.pho_id;


--
-- Name: adm_preferences; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.adm_preferences (
    prf_id integer NOT NULL,
    prf_org_id integer NOT NULL,
    prf_name character varying(50) NOT NULL,
    prf_value character varying(255)
);


ALTER TABLE public.adm_preferences OWNER TO postgres;

--
-- Name: adm_preferences_prf_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.adm_preferences_prf_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.adm_preferences_prf_id_seq OWNER TO postgres;

--
-- Name: adm_preferences_prf_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.adm_preferences_prf_id_seq OWNED BY public.adm_preferences.prf_id;


--
-- Name: adm_registrations; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.adm_registrations (
    reg_id integer NOT NULL,
    reg_org_id integer NOT NULL,
    reg_usr_id integer NOT NULL,
    reg_timestamp timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    reg_validation_id character varying(50)
);


ALTER TABLE public.adm_registrations OWNER TO postgres;

--
-- Name: adm_registrations_reg_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.adm_registrations_reg_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.adm_registrations_reg_id_seq OWNER TO postgres;

--
-- Name: adm_registrations_reg_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.adm_registrations_reg_id_seq OWNED BY public.adm_registrations.reg_id;


--
-- Name: adm_role_dependencies; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.adm_role_dependencies (
    rld_rol_id_parent integer NOT NULL,
    rld_rol_id_child integer NOT NULL,
    rld_comment text,
    rld_usr_id integer,
    rld_timestamp timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.adm_role_dependencies OWNER TO postgres;

--
-- Name: adm_roles; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.adm_roles (
    rol_id integer NOT NULL,
    rol_cat_id integer NOT NULL,
    rol_lst_id integer,
    rol_uuid character varying(36) NOT NULL,
    rol_name character varying(100) NOT NULL,
    rol_description character varying(4000),
    rol_assign_roles boolean DEFAULT false NOT NULL,
    rol_approve_users boolean DEFAULT false NOT NULL,
    rol_announcements boolean DEFAULT false NOT NULL,
    rol_events boolean DEFAULT false NOT NULL,
    rol_documents_files boolean DEFAULT false NOT NULL,
    rol_edit_user boolean DEFAULT false NOT NULL,
    rol_guestbook boolean DEFAULT false NOT NULL,
    rol_guestbook_comments boolean DEFAULT false NOT NULL,
    rol_mail_to_all boolean DEFAULT false NOT NULL,
    rol_mail_this_role smallint DEFAULT 0 NOT NULL,
    rol_photo boolean DEFAULT false NOT NULL,
    rol_profile boolean DEFAULT false NOT NULL,
    rol_weblinks boolean DEFAULT false NOT NULL,
    rol_all_lists_view boolean DEFAULT false NOT NULL,
    rol_default_registration boolean DEFAULT false NOT NULL,
    rol_leader_rights smallint DEFAULT 0 NOT NULL,
    rol_view_memberships smallint DEFAULT 0 NOT NULL,
    rol_view_members_profiles smallint DEFAULT 0 NOT NULL,
    rol_start_date date,
    rol_start_time time without time zone,
    rol_end_date date,
    rol_end_time time without time zone,
    rol_weekday smallint,
    rol_location character varying(100),
    rol_max_members integer,
    rol_cost double precision,
    rol_cost_period smallint,
    rol_usr_id_create integer,
    rol_timestamp_create timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    rol_usr_id_change integer,
    rol_timestamp_change timestamp without time zone,
    rol_valid boolean DEFAULT true NOT NULL,
    rol_system boolean DEFAULT false NOT NULL,
    rol_administrator boolean DEFAULT false NOT NULL
);


ALTER TABLE public.adm_roles OWNER TO postgres;

--
-- Name: adm_roles_rights; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.adm_roles_rights (
    ror_id integer NOT NULL,
    ror_name_intern character varying(50) NOT NULL,
    ror_table character varying(50) NOT NULL,
    ror_ror_id_parent integer
);


ALTER TABLE public.adm_roles_rights OWNER TO postgres;

--
-- Name: adm_roles_rights_data; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.adm_roles_rights_data (
    rrd_id integer NOT NULL,
    rrd_ror_id integer NOT NULL,
    rrd_rol_id integer NOT NULL,
    rrd_object_id integer NOT NULL,
    rrd_usr_id_create integer,
    rrd_timestamp_create timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.adm_roles_rights_data OWNER TO postgres;

--
-- Name: adm_roles_rights_data_rrd_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.adm_roles_rights_data_rrd_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.adm_roles_rights_data_rrd_id_seq OWNER TO postgres;

--
-- Name: adm_roles_rights_data_rrd_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.adm_roles_rights_data_rrd_id_seq OWNED BY public.adm_roles_rights_data.rrd_id;


--
-- Name: adm_roles_rights_ror_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.adm_roles_rights_ror_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.adm_roles_rights_ror_id_seq OWNER TO postgres;

--
-- Name: adm_roles_rights_ror_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.adm_roles_rights_ror_id_seq OWNED BY public.adm_roles_rights.ror_id;


--
-- Name: adm_roles_rol_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.adm_roles_rol_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.adm_roles_rol_id_seq OWNER TO postgres;

--
-- Name: adm_roles_rol_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.adm_roles_rol_id_seq OWNED BY public.adm_roles.rol_id;


--
-- Name: adm_rooms; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.adm_rooms (
    room_id integer NOT NULL,
    room_uuid character varying(36) NOT NULL,
    room_name character varying(50) NOT NULL,
    room_description text,
    room_capacity integer NOT NULL,
    room_overhang integer,
    room_usr_id_create integer,
    room_timestamp_create timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    room_usr_id_change integer,
    room_timestamp_change timestamp without time zone
);


ALTER TABLE public.adm_rooms OWNER TO postgres;

--
-- Name: adm_rooms_room_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.adm_rooms_room_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.adm_rooms_room_id_seq OWNER TO postgres;

--
-- Name: adm_rooms_room_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.adm_rooms_room_id_seq OWNED BY public.adm_rooms.room_id;


--
-- Name: adm_sessions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.adm_sessions (
    ses_id integer NOT NULL,
    ses_usr_id integer,
    ses_org_id integer NOT NULL,
    ses_session_id character varying(255) NOT NULL,
    ses_begin timestamp without time zone,
    ses_timestamp timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    ses_ip_address character varying(39) NOT NULL,
    ses_binary bytea,
    ses_reload boolean DEFAULT false NOT NULL
);


ALTER TABLE public.adm_sessions OWNER TO postgres;

--
-- Name: adm_sessions_ses_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.adm_sessions_ses_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.adm_sessions_ses_id_seq OWNER TO postgres;

--
-- Name: adm_sessions_ses_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.adm_sessions_ses_id_seq OWNED BY public.adm_sessions.ses_id;


--
-- Name: adm_statistics; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.adm_statistics (
    sta_id integer NOT NULL,
    sta_org_id integer NOT NULL,
    sta_name character varying(50) NOT NULL,
    sta_title character varying(200),
    sta_subtitle character varying(200),
    sta_std_role integer NOT NULL
);


ALTER TABLE public.adm_statistics OWNER TO postgres;

--
-- Name: adm_statistics_columns; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.adm_statistics_columns (
    stc_id integer NOT NULL,
    stc_label character varying(50),
    stc_field_condition character varying(200),
    stc_profile_field character varying(50),
    stc_function_main character varying(50),
    stc_function_arg character varying(200),
    stc_function_total character varying(50),
    stc_stt_id integer NOT NULL
);


ALTER TABLE public.adm_statistics_columns OWNER TO postgres;

--
-- Name: adm_statistics_columns_stc_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.adm_statistics_columns_stc_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.adm_statistics_columns_stc_id_seq OWNER TO postgres;

--
-- Name: adm_statistics_columns_stc_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.adm_statistics_columns_stc_id_seq OWNED BY public.adm_statistics_columns.stc_id;


--
-- Name: adm_statistics_rows; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.adm_statistics_rows (
    str_id integer NOT NULL,
    str_label character varying(50),
    str_field_condition character varying(200),
    str_profile_field character varying(50),
    str_stt_id integer NOT NULL
);


ALTER TABLE public.adm_statistics_rows OWNER TO postgres;

--
-- Name: adm_statistics_rows_str_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.adm_statistics_rows_str_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.adm_statistics_rows_str_id_seq OWNER TO postgres;

--
-- Name: adm_statistics_rows_str_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.adm_statistics_rows_str_id_seq OWNED BY public.adm_statistics_rows.str_id;


--
-- Name: adm_statistics_sta_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.adm_statistics_sta_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.adm_statistics_sta_id_seq OWNER TO postgres;

--
-- Name: adm_statistics_sta_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.adm_statistics_sta_id_seq OWNED BY public.adm_statistics.sta_id;


--
-- Name: adm_statistics_tables; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.adm_statistics_tables (
    stt_id integer NOT NULL,
    stt_title character varying(200),
    stt_role integer,
    stt_first_column_label character varying(50),
    stt_sta_id integer NOT NULL
);


ALTER TABLE public.adm_statistics_tables OWNER TO postgres;

--
-- Name: adm_statistics_tables_stt_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.adm_statistics_tables_stt_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.adm_statistics_tables_stt_id_seq OWNER TO postgres;

--
-- Name: adm_statistics_tables_stt_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.adm_statistics_tables_stt_id_seq OWNED BY public.adm_statistics_tables.stt_id;


--
-- Name: adm_texts; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.adm_texts (
    txt_id integer NOT NULL,
    txt_org_id integer NOT NULL,
    txt_name character varying(100) NOT NULL,
    txt_text text
);


ALTER TABLE public.adm_texts OWNER TO postgres;

--
-- Name: adm_texts_txt_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.adm_texts_txt_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.adm_texts_txt_id_seq OWNER TO postgres;

--
-- Name: adm_texts_txt_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.adm_texts_txt_id_seq OWNED BY public.adm_texts.txt_id;


--
-- Name: adm_user_data; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.adm_user_data (
    usd_id integer NOT NULL,
    usd_usr_id integer NOT NULL,
    usd_usf_id integer NOT NULL,
    usd_value character varying(4000)
);


ALTER TABLE public.adm_user_data OWNER TO postgres;

--
-- Name: adm_user_data_usd_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.adm_user_data_usd_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.adm_user_data_usd_id_seq OWNER TO postgres;

--
-- Name: adm_user_data_usd_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.adm_user_data_usd_id_seq OWNED BY public.adm_user_data.usd_id;


--
-- Name: adm_user_fields; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.adm_user_fields (
    usf_id integer NOT NULL,
    usf_cat_id integer NOT NULL,
    usf_uuid character varying(36) NOT NULL,
    usf_type character varying(30) NOT NULL,
    usf_name_intern character varying(110) NOT NULL,
    usf_name character varying(100) NOT NULL,
    usf_description text,
    usf_description_inline boolean DEFAULT false NOT NULL,
    usf_value_list text,
    usf_default_value character varying(100),
    usf_regex character varying(100),
    usf_icon character varying(100),
    usf_url character varying(2000),
    usf_system boolean DEFAULT false NOT NULL,
    usf_disabled boolean DEFAULT false NOT NULL,
    usf_hidden boolean DEFAULT false NOT NULL,
    usf_registration boolean DEFAULT false NOT NULL,
    usf_required_input smallint DEFAULT 0 NOT NULL,
    usf_sequence smallint NOT NULL,
    usf_usr_id_create integer,
    usf_timestamp_create timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    usf_usr_id_change integer,
    usf_timestamp_change timestamp without time zone
);


ALTER TABLE public.adm_user_fields OWNER TO postgres;

--
-- Name: adm_user_fields_usf_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.adm_user_fields_usf_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.adm_user_fields_usf_id_seq OWNER TO postgres;

--
-- Name: adm_user_fields_usf_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.adm_user_fields_usf_id_seq OWNED BY public.adm_user_fields.usf_id;


--
-- Name: adm_user_log; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.adm_user_log (
    usl_id integer NOT NULL,
    usl_usr_id integer NOT NULL,
    usl_usf_id integer NOT NULL,
    usl_value_old character varying(4000),
    usl_value_new character varying(4000),
    usl_usr_id_create integer,
    usl_timestamp_create timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    usl_comment character varying(255)
);


ALTER TABLE public.adm_user_log OWNER TO postgres;

--
-- Name: adm_user_log_usl_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.adm_user_log_usl_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.adm_user_log_usl_id_seq OWNER TO postgres;

--
-- Name: adm_user_log_usl_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.adm_user_log_usl_id_seq OWNED BY public.adm_user_log.usl_id;


--
-- Name: adm_user_relation_types; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.adm_user_relation_types (
    urt_id integer NOT NULL,
    urt_uuid character varying(36) NOT NULL,
    urt_name character varying(100) NOT NULL,
    urt_name_male character varying(100) NOT NULL,
    urt_name_female character varying(100) NOT NULL,
    urt_edit_user boolean DEFAULT false NOT NULL,
    urt_id_inverse integer,
    urt_usr_id_create integer,
    urt_timestamp_create timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    urt_usr_id_change integer,
    urt_timestamp_change timestamp without time zone
);


ALTER TABLE public.adm_user_relation_types OWNER TO postgres;

--
-- Name: adm_user_relation_types_urt_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.adm_user_relation_types_urt_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.adm_user_relation_types_urt_id_seq OWNER TO postgres;

--
-- Name: adm_user_relation_types_urt_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.adm_user_relation_types_urt_id_seq OWNED BY public.adm_user_relation_types.urt_id;


--
-- Name: adm_user_relations; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.adm_user_relations (
    ure_id integer NOT NULL,
    ure_urt_id integer NOT NULL,
    ure_usr_id1 integer NOT NULL,
    ure_usr_id2 integer NOT NULL,
    ure_usr_id_create integer,
    ure_timestamp_create timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    ure_usr_id_change integer,
    ure_timestamp_change timestamp without time zone
);


ALTER TABLE public.adm_user_relations OWNER TO postgres;

--
-- Name: adm_user_relations_ure_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.adm_user_relations_ure_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.adm_user_relations_ure_id_seq OWNER TO postgres;

--
-- Name: adm_user_relations_ure_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.adm_user_relations_ure_id_seq OWNED BY public.adm_user_relations.ure_id;


--
-- Name: adm_users; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.adm_users (
    usr_id integer NOT NULL,
    usr_uuid character varying(36) NOT NULL,
    usr_login_name character varying(254),
    usr_password character varying(255),
    usr_photo bytea,
    usr_text text,
    usr_pw_reset_id character varying(50),
    usr_pw_reset_timestamp timestamp without time zone,
    usr_last_login timestamp without time zone,
    usr_actual_login timestamp without time zone,
    usr_number_login integer DEFAULT 0 NOT NULL,
    usr_date_invalid timestamp without time zone,
    usr_number_invalid smallint DEFAULT 0 NOT NULL,
    usr_usr_id_create integer,
    usr_timestamp_create timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    usr_usr_id_change integer,
    usr_timestamp_change timestamp without time zone,
    usr_valid boolean DEFAULT false NOT NULL
);


ALTER TABLE public.adm_users OWNER TO postgres;

--
-- Name: adm_users_usr_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.adm_users_usr_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.adm_users_usr_id_seq OWNER TO postgres;

--
-- Name: adm_users_usr_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.adm_users_usr_id_seq OWNED BY public.adm_users.usr_id;


--
-- Name: membership_sales; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.membership_sales (
    ms_id integer NOT NULL,
    ms_payment_service character varying(36) NOT NULL,
    ms_payment_status character varying(20) NOT NULL,
    ms_payment_id character varying(200),
    ms_membership_year integer NOT NULL,
    ms_usr1_id integer NOT NULL,
    ms_usr1_fee real NOT NULL,
    ms_usr1_friend boolean NOT NULL,
    ms_usr1_friend_fee real,
    ms_usr2_id integer,
    ms_usr2_fee real NOT NULL,
    ms_usr2_friend boolean NOT NULL,
    ms_usr2_friend_fee real NOT NULL,
    ms_donation real NOT NULL,
    ms_donation_museum real NOT NULL,
    timestamp_create timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    ms_giftaid boolean DEFAULT false NOT NULL
);


ALTER TABLE public.membership_sales OWNER TO postgres;

--
-- Name: membership_sales_ms_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.membership_sales_ms_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.membership_sales_ms_id_seq OWNER TO postgres;

--
-- Name: membership_sales_ms_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.membership_sales_ms_id_seq OWNED BY public.membership_sales.ms_id;


--
-- Name: adm_announcements ann_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_announcements ALTER COLUMN ann_id SET DEFAULT nextval('public.adm_announcements_ann_id_seq'::regclass);


--
-- Name: adm_auto_login atl_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_auto_login ALTER COLUMN atl_id SET DEFAULT nextval('public.adm_auto_login_atl_id_seq'::regclass);


--
-- Name: adm_categories cat_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_categories ALTER COLUMN cat_id SET DEFAULT nextval('public.adm_categories_cat_id_seq'::regclass);


--
-- Name: adm_category_report crt_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_category_report ALTER COLUMN crt_id SET DEFAULT nextval('public.adm_category_report_crt_id_seq'::regclass);


--
-- Name: adm_components com_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_components ALTER COLUMN com_id SET DEFAULT nextval('public.adm_components_com_id_seq'::regclass);


--
-- Name: adm_events dat_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_events ALTER COLUMN dat_id SET DEFAULT nextval('public.adm_events_dat_id_seq'::regclass);


--
-- Name: adm_files fil_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_files ALTER COLUMN fil_id SET DEFAULT nextval('public.adm_files_fil_id_seq'::regclass);


--
-- Name: adm_folders fol_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_folders ALTER COLUMN fol_id SET DEFAULT nextval('public.adm_folders_fol_id_seq'::regclass);


--
-- Name: adm_guestbook gbo_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_guestbook ALTER COLUMN gbo_id SET DEFAULT nextval('public.adm_guestbook_gbo_id_seq'::regclass);


--
-- Name: adm_guestbook_comments gbc_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_guestbook_comments ALTER COLUMN gbc_id SET DEFAULT nextval('public.adm_guestbook_comments_gbc_id_seq'::regclass);


--
-- Name: adm_links lnk_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_links ALTER COLUMN lnk_id SET DEFAULT nextval('public.adm_links_lnk_id_seq'::regclass);


--
-- Name: adm_list_columns lsc_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_list_columns ALTER COLUMN lsc_id SET DEFAULT nextval('public.adm_list_columns_lsc_id_seq'::regclass);


--
-- Name: adm_lists lst_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_lists ALTER COLUMN lst_id SET DEFAULT nextval('public.adm_lists_lst_id_seq'::regclass);


--
-- Name: adm_members mem_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_members ALTER COLUMN mem_id SET DEFAULT nextval('public.adm_members_mem_id_seq'::regclass);


--
-- Name: adm_members_interests mi_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_members_interests ALTER COLUMN mi_id SET DEFAULT nextval('public.adm_members_interests_mi_id_seq'::regclass);


--
-- Name: adm_members_other_interests moi_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_members_other_interests ALTER COLUMN moi_id SET DEFAULT nextval('public.adm_members_other_interests_moi_id_seq'::regclass);


--
-- Name: adm_menu men_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_menu ALTER COLUMN men_id SET DEFAULT nextval('public.adm_menu_men_id_seq'::regclass);


--
-- Name: adm_messages msg_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_messages ALTER COLUMN msg_id SET DEFAULT nextval('public.adm_messages_msg_id_seq'::regclass);


--
-- Name: adm_messages_attachments msa_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_messages_attachments ALTER COLUMN msa_id SET DEFAULT nextval('public.adm_messages_attachments_msa_id_seq'::regclass);


--
-- Name: adm_messages_content msc_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_messages_content ALTER COLUMN msc_id SET DEFAULT nextval('public.adm_messages_content_msc_id_seq'::regclass);


--
-- Name: adm_messages_recipients msr_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_messages_recipients ALTER COLUMN msr_id SET DEFAULT nextval('public.adm_messages_recipients_msr_id_seq'::regclass);


--
-- Name: adm_organizations org_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_organizations ALTER COLUMN org_id SET DEFAULT nextval('public.adm_organizations_org_id_seq'::regclass);


--
-- Name: adm_photos pho_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_photos ALTER COLUMN pho_id SET DEFAULT nextval('public.adm_photos_pho_id_seq'::regclass);


--
-- Name: adm_preferences prf_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_preferences ALTER COLUMN prf_id SET DEFAULT nextval('public.adm_preferences_prf_id_seq'::regclass);


--
-- Name: adm_registrations reg_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_registrations ALTER COLUMN reg_id SET DEFAULT nextval('public.adm_registrations_reg_id_seq'::regclass);


--
-- Name: adm_roles rol_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_roles ALTER COLUMN rol_id SET DEFAULT nextval('public.adm_roles_rol_id_seq'::regclass);


--
-- Name: adm_roles_rights ror_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_roles_rights ALTER COLUMN ror_id SET DEFAULT nextval('public.adm_roles_rights_ror_id_seq'::regclass);


--
-- Name: adm_roles_rights_data rrd_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_roles_rights_data ALTER COLUMN rrd_id SET DEFAULT nextval('public.adm_roles_rights_data_rrd_id_seq'::regclass);


--
-- Name: adm_rooms room_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_rooms ALTER COLUMN room_id SET DEFAULT nextval('public.adm_rooms_room_id_seq'::regclass);


--
-- Name: adm_sessions ses_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_sessions ALTER COLUMN ses_id SET DEFAULT nextval('public.adm_sessions_ses_id_seq'::regclass);


--
-- Name: adm_statistics sta_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_statistics ALTER COLUMN sta_id SET DEFAULT nextval('public.adm_statistics_sta_id_seq'::regclass);


--
-- Name: adm_statistics_columns stc_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_statistics_columns ALTER COLUMN stc_id SET DEFAULT nextval('public.adm_statistics_columns_stc_id_seq'::regclass);


--
-- Name: adm_statistics_rows str_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_statistics_rows ALTER COLUMN str_id SET DEFAULT nextval('public.adm_statistics_rows_str_id_seq'::regclass);


--
-- Name: adm_statistics_tables stt_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_statistics_tables ALTER COLUMN stt_id SET DEFAULT nextval('public.adm_statistics_tables_stt_id_seq'::regclass);


--
-- Name: adm_texts txt_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_texts ALTER COLUMN txt_id SET DEFAULT nextval('public.adm_texts_txt_id_seq'::regclass);


--
-- Name: adm_user_data usd_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_user_data ALTER COLUMN usd_id SET DEFAULT nextval('public.adm_user_data_usd_id_seq'::regclass);


--
-- Name: adm_user_fields usf_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_user_fields ALTER COLUMN usf_id SET DEFAULT nextval('public.adm_user_fields_usf_id_seq'::regclass);


--
-- Name: adm_user_log usl_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_user_log ALTER COLUMN usl_id SET DEFAULT nextval('public.adm_user_log_usl_id_seq'::regclass);


--
-- Name: adm_user_relation_types urt_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_user_relation_types ALTER COLUMN urt_id SET DEFAULT nextval('public.adm_user_relation_types_urt_id_seq'::regclass);


--
-- Name: adm_user_relations ure_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_user_relations ALTER COLUMN ure_id SET DEFAULT nextval('public.adm_user_relations_ure_id_seq'::regclass);


--
-- Name: adm_users usr_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_users ALTER COLUMN usr_id SET DEFAULT nextval('public.adm_users_usr_id_seq'::regclass);


--
-- Name: membership_sales ms_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.membership_sales ALTER COLUMN ms_id SET DEFAULT nextval('public.membership_sales_ms_id_seq'::regclass);


--
-- Name: adm_announcements adm_announcements_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_announcements
    ADD CONSTRAINT adm_announcements_pkey PRIMARY KEY (ann_id);


--
-- Name: adm_auto_login adm_auto_login_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_auto_login
    ADD CONSTRAINT adm_auto_login_pkey PRIMARY KEY (atl_id);


--
-- Name: adm_categories adm_categories_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_categories
    ADD CONSTRAINT adm_categories_pkey PRIMARY KEY (cat_id);


--
-- Name: adm_category_report adm_category_report_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_category_report
    ADD CONSTRAINT adm_category_report_pkey PRIMARY KEY (crt_id);


--
-- Name: adm_components adm_components_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_components
    ADD CONSTRAINT adm_components_pkey PRIMARY KEY (com_id);


--
-- Name: adm_events adm_events_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_events
    ADD CONSTRAINT adm_events_pkey PRIMARY KEY (dat_id);


--
-- Name: adm_files adm_files_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_files
    ADD CONSTRAINT adm_files_pkey PRIMARY KEY (fil_id);


--
-- Name: adm_folders adm_folders_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_folders
    ADD CONSTRAINT adm_folders_pkey PRIMARY KEY (fol_id);


--
-- Name: adm_guestbook_comments adm_guestbook_comments_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_guestbook_comments
    ADD CONSTRAINT adm_guestbook_comments_pkey PRIMARY KEY (gbc_id);


--
-- Name: adm_guestbook adm_guestbook_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_guestbook
    ADD CONSTRAINT adm_guestbook_pkey PRIMARY KEY (gbo_id);


--
-- Name: adm_interests adm_interests_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_interests
    ADD CONSTRAINT adm_interests_pkey PRIMARY KEY (ntrst_id);


--
-- Name: adm_links adm_links_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_links
    ADD CONSTRAINT adm_links_pkey PRIMARY KEY (lnk_id);


--
-- Name: adm_list_columns adm_list_columns_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_list_columns
    ADD CONSTRAINT adm_list_columns_pkey PRIMARY KEY (lsc_id);


--
-- Name: adm_lists adm_lists_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_lists
    ADD CONSTRAINT adm_lists_pkey PRIMARY KEY (lst_id);


--
-- Name: adm_members_interests adm_members_interests_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_members_interests
    ADD CONSTRAINT adm_members_interests_pkey PRIMARY KEY (mi_id);


--
-- Name: adm_members_other_interests adm_members_other_interests_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_members_other_interests
    ADD CONSTRAINT adm_members_other_interests_pkey PRIMARY KEY (moi_id);


--
-- Name: adm_members adm_members_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_members
    ADD CONSTRAINT adm_members_pkey PRIMARY KEY (mem_id);


--
-- Name: adm_menu adm_menu_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_menu
    ADD CONSTRAINT adm_menu_pkey PRIMARY KEY (men_id);


--
-- Name: adm_messages_attachments adm_messages_attachments_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_messages_attachments
    ADD CONSTRAINT adm_messages_attachments_pkey PRIMARY KEY (msa_id);


--
-- Name: adm_messages_content adm_messages_content_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_messages_content
    ADD CONSTRAINT adm_messages_content_pkey PRIMARY KEY (msc_id);


--
-- Name: adm_messages adm_messages_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_messages
    ADD CONSTRAINT adm_messages_pkey PRIMARY KEY (msg_id);


--
-- Name: adm_messages_recipients adm_messages_recipients_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_messages_recipients
    ADD CONSTRAINT adm_messages_recipients_pkey PRIMARY KEY (msr_id);


--
-- Name: adm_organizations adm_organizations_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_organizations
    ADD CONSTRAINT adm_organizations_pkey PRIMARY KEY (org_id);


--
-- Name: adm_photos adm_photos_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_photos
    ADD CONSTRAINT adm_photos_pkey PRIMARY KEY (pho_id);


--
-- Name: adm_preferences adm_preferences_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_preferences
    ADD CONSTRAINT adm_preferences_pkey PRIMARY KEY (prf_id);


--
-- Name: adm_registrations adm_registrations_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_registrations
    ADD CONSTRAINT adm_registrations_pkey PRIMARY KEY (reg_id);


--
-- Name: adm_role_dependencies adm_role_dependencies_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_role_dependencies
    ADD CONSTRAINT adm_role_dependencies_pkey PRIMARY KEY (rld_rol_id_parent, rld_rol_id_child);


--
-- Name: adm_roles adm_roles_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_roles
    ADD CONSTRAINT adm_roles_pkey PRIMARY KEY (rol_id);


--
-- Name: adm_roles_rights_data adm_roles_rights_data_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_roles_rights_data
    ADD CONSTRAINT adm_roles_rights_data_pkey PRIMARY KEY (rrd_id);


--
-- Name: adm_roles_rights adm_roles_rights_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_roles_rights
    ADD CONSTRAINT adm_roles_rights_pkey PRIMARY KEY (ror_id);


--
-- Name: adm_rooms adm_rooms_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_rooms
    ADD CONSTRAINT adm_rooms_pkey PRIMARY KEY (room_id);


--
-- Name: adm_sessions adm_sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_sessions
    ADD CONSTRAINT adm_sessions_pkey PRIMARY KEY (ses_id);


--
-- Name: adm_statistics_columns adm_statistics_columns_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_statistics_columns
    ADD CONSTRAINT adm_statistics_columns_pkey PRIMARY KEY (stc_id);


--
-- Name: adm_statistics adm_statistics_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_statistics
    ADD CONSTRAINT adm_statistics_pkey PRIMARY KEY (sta_id);


--
-- Name: adm_statistics_rows adm_statistics_rows_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_statistics_rows
    ADD CONSTRAINT adm_statistics_rows_pkey PRIMARY KEY (str_id);


--
-- Name: adm_statistics_tables adm_statistics_tables_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_statistics_tables
    ADD CONSTRAINT adm_statistics_tables_pkey PRIMARY KEY (stt_id);


--
-- Name: adm_texts adm_texts_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_texts
    ADD CONSTRAINT adm_texts_pkey PRIMARY KEY (txt_id);


--
-- Name: adm_user_data adm_user_data_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_user_data
    ADD CONSTRAINT adm_user_data_pkey PRIMARY KEY (usd_id);


--
-- Name: adm_user_fields adm_user_fields_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_user_fields
    ADD CONSTRAINT adm_user_fields_pkey PRIMARY KEY (usf_id);


--
-- Name: adm_user_log adm_user_log_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_user_log
    ADD CONSTRAINT adm_user_log_pkey PRIMARY KEY (usl_id);


--
-- Name: adm_user_relation_types adm_user_relation_types_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_user_relation_types
    ADD CONSTRAINT adm_user_relation_types_pkey PRIMARY KEY (urt_id);


--
-- Name: adm_user_relations adm_user_relations_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_user_relations
    ADD CONSTRAINT adm_user_relations_pkey PRIMARY KEY (ure_id);


--
-- Name: adm_users adm_users_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_users
    ADD CONSTRAINT adm_users_pkey PRIMARY KEY (usr_id);


--
-- Name: membership_sales membership_sales_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.membership_sales
    ADD CONSTRAINT membership_sales_pkey PRIMARY KEY (ms_id);


--
-- Name: adm_idx_ann_uuid; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX adm_idx_ann_uuid ON public.adm_announcements USING btree (ann_uuid);


--
-- Name: adm_idx_cat_uuid; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX adm_idx_cat_uuid ON public.adm_categories USING btree (cat_uuid);


--
-- Name: adm_idx_dat_uuid; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX adm_idx_dat_uuid ON public.adm_events USING btree (dat_uuid);


--
-- Name: adm_idx_fil_uuid; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX adm_idx_fil_uuid ON public.adm_files USING btree (fil_uuid);


--
-- Name: adm_idx_fol_uuid; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX adm_idx_fol_uuid ON public.adm_folders USING btree (fol_uuid);


--
-- Name: adm_idx_gbc_uuid; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX adm_idx_gbc_uuid ON public.adm_guestbook_comments USING btree (gbc_uuid);


--
-- Name: adm_idx_gbo_uuid; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX adm_idx_gbo_uuid ON public.adm_guestbook USING btree (gbo_uuid);


--
-- Name: adm_idx_lnk_uuid; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX adm_idx_lnk_uuid ON public.adm_links USING btree (lnk_uuid);


--
-- Name: adm_idx_lst_uuid; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX adm_idx_lst_uuid ON public.adm_lists USING btree (lst_uuid);


--
-- Name: adm_idx_mem_rol_usr_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX adm_idx_mem_rol_usr_id ON public.adm_members USING btree (mem_rol_id, mem_usr_id);


--
-- Name: adm_idx_mem_uuid; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX adm_idx_mem_uuid ON public.adm_members USING btree (mem_uuid);


--
-- Name: adm_idx_men_men_id_parent; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX adm_idx_men_men_id_parent ON public.adm_menu USING btree (men_men_id_parent);


--
-- Name: adm_idx_men_uuid; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX adm_idx_men_uuid ON public.adm_menu USING btree (men_uuid);


--
-- Name: adm_idx_msg_uuid; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX adm_idx_msg_uuid ON public.adm_messages USING btree (msg_uuid);


--
-- Name: adm_idx_org_shortname; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX adm_idx_org_shortname ON public.adm_organizations USING btree (org_shortname);


--
-- Name: adm_idx_org_uuid; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX adm_idx_org_uuid ON public.adm_organizations USING btree (org_uuid);


--
-- Name: adm_idx_pho_uuid; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX adm_idx_pho_uuid ON public.adm_photos USING btree (pho_uuid);


--
-- Name: adm_idx_prf_org_id_name; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX adm_idx_prf_org_id_name ON public.adm_preferences USING btree (prf_org_id, prf_name);


--
-- Name: adm_idx_rol_uuid; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX adm_idx_rol_uuid ON public.adm_roles USING btree (rol_uuid);


--
-- Name: adm_idx_room_uuid; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX adm_idx_room_uuid ON public.adm_rooms USING btree (room_uuid);


--
-- Name: adm_idx_rrd_ror_rol_object_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX adm_idx_rrd_ror_rol_object_id ON public.adm_roles_rights_data USING btree (rrd_ror_id, rrd_rol_id, rrd_object_id);


--
-- Name: adm_idx_session_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX adm_idx_session_id ON public.adm_sessions USING btree (ses_session_id);


--
-- Name: adm_idx_ure_urt_name; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX adm_idx_ure_urt_name ON public.adm_user_relation_types USING btree (urt_name);


--
-- Name: adm_idx_ure_urt_usr; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX adm_idx_ure_urt_usr ON public.adm_user_relations USING btree (ure_urt_id, ure_usr_id1, ure_usr_id2);


--
-- Name: adm_idx_urt_uuid; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX adm_idx_urt_uuid ON public.adm_user_relation_types USING btree (urt_uuid);


--
-- Name: adm_idx_usd_usr_usf_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX adm_idx_usd_usr_usf_id ON public.adm_user_data USING btree (usd_usr_id, usd_usf_id);


--
-- Name: adm_idx_usf_name_intern; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX adm_idx_usf_name_intern ON public.adm_user_fields USING btree (usf_name_intern);


--
-- Name: adm_idx_usf_uuid; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX adm_idx_usf_uuid ON public.adm_user_fields USING btree (usf_uuid);


--
-- Name: adm_idx_usr_login_name; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX adm_idx_usr_login_name ON public.adm_users USING btree (usr_login_name);


--
-- Name: adm_idx_usr_uuid; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX adm_idx_usr_uuid ON public.adm_users USING btree (usr_uuid);


--
-- Name: adm_announcements adm_fk_ann_cat; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_announcements
    ADD CONSTRAINT adm_fk_ann_cat FOREIGN KEY (ann_cat_id) REFERENCES public.adm_categories(cat_id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: adm_announcements adm_fk_ann_usr_change; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_announcements
    ADD CONSTRAINT adm_fk_ann_usr_change FOREIGN KEY (ann_usr_id_change) REFERENCES public.adm_users(usr_id) ON UPDATE RESTRICT ON DELETE SET NULL;


--
-- Name: adm_announcements adm_fk_ann_usr_create; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_announcements
    ADD CONSTRAINT adm_fk_ann_usr_create FOREIGN KEY (ann_usr_id_create) REFERENCES public.adm_users(usr_id) ON UPDATE RESTRICT ON DELETE SET NULL;


--
-- Name: adm_auto_login adm_fk_atl_org; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_auto_login
    ADD CONSTRAINT adm_fk_atl_org FOREIGN KEY (atl_org_id) REFERENCES public.adm_organizations(org_id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: adm_auto_login adm_fk_atl_usr; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_auto_login
    ADD CONSTRAINT adm_fk_atl_usr FOREIGN KEY (atl_usr_id) REFERENCES public.adm_users(usr_id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: adm_categories adm_fk_cat_org; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_categories
    ADD CONSTRAINT adm_fk_cat_org FOREIGN KEY (cat_org_id) REFERENCES public.adm_organizations(org_id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: adm_categories adm_fk_cat_usr_change; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_categories
    ADD CONSTRAINT adm_fk_cat_usr_change FOREIGN KEY (cat_usr_id_change) REFERENCES public.adm_users(usr_id) ON UPDATE RESTRICT ON DELETE SET NULL;


--
-- Name: adm_categories adm_fk_cat_usr_create; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_categories
    ADD CONSTRAINT adm_fk_cat_usr_create FOREIGN KEY (cat_usr_id_create) REFERENCES public.adm_users(usr_id) ON UPDATE RESTRICT ON DELETE SET NULL;


--
-- Name: adm_category_report adm_fk_crt_org; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_category_report
    ADD CONSTRAINT adm_fk_crt_org FOREIGN KEY (crt_org_id) REFERENCES public.adm_organizations(org_id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: adm_events adm_fk_dat_cat; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_events
    ADD CONSTRAINT adm_fk_dat_cat FOREIGN KEY (dat_cat_id) REFERENCES public.adm_categories(cat_id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: adm_events adm_fk_dat_rol; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_events
    ADD CONSTRAINT adm_fk_dat_rol FOREIGN KEY (dat_rol_id) REFERENCES public.adm_roles(rol_id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: adm_events adm_fk_dat_room; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_events
    ADD CONSTRAINT adm_fk_dat_room FOREIGN KEY (dat_room_id) REFERENCES public.adm_rooms(room_id) ON UPDATE RESTRICT ON DELETE SET NULL;


--
-- Name: adm_events adm_fk_dat_usr_change; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_events
    ADD CONSTRAINT adm_fk_dat_usr_change FOREIGN KEY (dat_usr_id_change) REFERENCES public.adm_users(usr_id) ON UPDATE RESTRICT ON DELETE SET NULL;


--
-- Name: adm_events adm_fk_dat_usr_create; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_events
    ADD CONSTRAINT adm_fk_dat_usr_create FOREIGN KEY (dat_usr_id_create) REFERENCES public.adm_users(usr_id) ON UPDATE RESTRICT ON DELETE SET NULL;


--
-- Name: adm_files adm_fk_fil_fol; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_files
    ADD CONSTRAINT adm_fk_fil_fol FOREIGN KEY (fil_fol_id) REFERENCES public.adm_folders(fol_id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: adm_files adm_fk_fil_usr; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_files
    ADD CONSTRAINT adm_fk_fil_usr FOREIGN KEY (fil_usr_id) REFERENCES public.adm_users(usr_id) ON UPDATE RESTRICT ON DELETE SET NULL;


--
-- Name: adm_folders adm_fk_fol_fol_parent; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_folders
    ADD CONSTRAINT adm_fk_fol_fol_parent FOREIGN KEY (fol_fol_id_parent) REFERENCES public.adm_folders(fol_id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: adm_folders adm_fk_fol_org; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_folders
    ADD CONSTRAINT adm_fk_fol_org FOREIGN KEY (fol_org_id) REFERENCES public.adm_organizations(org_id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: adm_folders adm_fk_fol_usr; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_folders
    ADD CONSTRAINT adm_fk_fol_usr FOREIGN KEY (fol_usr_id) REFERENCES public.adm_users(usr_id) ON UPDATE RESTRICT ON DELETE SET NULL;


--
-- Name: adm_guestbook_comments adm_fk_gbc_gbo; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_guestbook_comments
    ADD CONSTRAINT adm_fk_gbc_gbo FOREIGN KEY (gbc_gbo_id) REFERENCES public.adm_guestbook(gbo_id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: adm_guestbook_comments adm_fk_gbc_usr_change; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_guestbook_comments
    ADD CONSTRAINT adm_fk_gbc_usr_change FOREIGN KEY (gbc_usr_id_change) REFERENCES public.adm_users(usr_id) ON UPDATE RESTRICT ON DELETE SET NULL;


--
-- Name: adm_guestbook_comments adm_fk_gbc_usr_create; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_guestbook_comments
    ADD CONSTRAINT adm_fk_gbc_usr_create FOREIGN KEY (gbc_usr_id_create) REFERENCES public.adm_users(usr_id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: adm_guestbook adm_fk_gbo_org; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_guestbook
    ADD CONSTRAINT adm_fk_gbo_org FOREIGN KEY (gbo_org_id) REFERENCES public.adm_organizations(org_id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: adm_guestbook adm_fk_gbo_usr_change; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_guestbook
    ADD CONSTRAINT adm_fk_gbo_usr_change FOREIGN KEY (gbo_usr_id_change) REFERENCES public.adm_users(usr_id) ON UPDATE RESTRICT ON DELETE SET NULL;


--
-- Name: adm_guestbook adm_fk_gbo_usr_create; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_guestbook
    ADD CONSTRAINT adm_fk_gbo_usr_create FOREIGN KEY (gbo_usr_id_create) REFERENCES public.adm_users(usr_id) ON UPDATE RESTRICT ON DELETE SET NULL;


--
-- Name: adm_ids adm_fk_ids_usr_id; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_ids
    ADD CONSTRAINT adm_fk_ids_usr_id FOREIGN KEY (ids_usr_id) REFERENCES public.adm_users(usr_id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: adm_links adm_fk_lnk_cat; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_links
    ADD CONSTRAINT adm_fk_lnk_cat FOREIGN KEY (lnk_cat_id) REFERENCES public.adm_categories(cat_id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: adm_links adm_fk_lnk_usr_change; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_links
    ADD CONSTRAINT adm_fk_lnk_usr_change FOREIGN KEY (lnk_usr_id_change) REFERENCES public.adm_users(usr_id) ON UPDATE RESTRICT ON DELETE SET NULL;


--
-- Name: adm_links adm_fk_lnk_usr_create; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_links
    ADD CONSTRAINT adm_fk_lnk_usr_create FOREIGN KEY (lnk_usr_id_create) REFERENCES public.adm_users(usr_id) ON UPDATE RESTRICT ON DELETE SET NULL;


--
-- Name: adm_list_columns adm_fk_lsc_lst; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_list_columns
    ADD CONSTRAINT adm_fk_lsc_lst FOREIGN KEY (lsc_lst_id) REFERENCES public.adm_lists(lst_id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: adm_list_columns adm_fk_lsc_usf; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_list_columns
    ADD CONSTRAINT adm_fk_lsc_usf FOREIGN KEY (lsc_usf_id) REFERENCES public.adm_user_fields(usf_id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: adm_lists adm_fk_lst_org; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_lists
    ADD CONSTRAINT adm_fk_lst_org FOREIGN KEY (lst_org_id) REFERENCES public.adm_organizations(org_id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: adm_lists adm_fk_lst_usr; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_lists
    ADD CONSTRAINT adm_fk_lst_usr FOREIGN KEY (lst_usr_id) REFERENCES public.adm_users(usr_id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: adm_members adm_fk_mem_rol; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_members
    ADD CONSTRAINT adm_fk_mem_rol FOREIGN KEY (mem_rol_id) REFERENCES public.adm_roles(rol_id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: adm_members adm_fk_mem_usr; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_members
    ADD CONSTRAINT adm_fk_mem_usr FOREIGN KEY (mem_usr_id) REFERENCES public.adm_users(usr_id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: adm_members adm_fk_mem_usr_change; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_members
    ADD CONSTRAINT adm_fk_mem_usr_change FOREIGN KEY (mem_usr_id_change) REFERENCES public.adm_users(usr_id) ON UPDATE RESTRICT ON DELETE SET NULL;


--
-- Name: adm_members adm_fk_mem_usr_create; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_members
    ADD CONSTRAINT adm_fk_mem_usr_create FOREIGN KEY (mem_usr_id_create) REFERENCES public.adm_users(usr_id) ON UPDATE RESTRICT ON DELETE SET NULL;


--
-- Name: adm_menu adm_fk_men_com_id; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_menu
    ADD CONSTRAINT adm_fk_men_com_id FOREIGN KEY (men_com_id) REFERENCES public.adm_components(com_id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: adm_menu adm_fk_men_men_parent; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_menu
    ADD CONSTRAINT adm_fk_men_men_parent FOREIGN KEY (men_men_id_parent) REFERENCES public.adm_menu(men_id) ON UPDATE RESTRICT ON DELETE SET NULL;


--
-- Name: adm_members_interests adm_fk_mi_interest; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_members_interests
    ADD CONSTRAINT adm_fk_mi_interest FOREIGN KEY (mi_interest_id) REFERENCES public.adm_interests(ntrst_id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: adm_members_interests adm_fk_mi_usr; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_members_interests
    ADD CONSTRAINT adm_fk_mi_usr FOREIGN KEY (mi_usr_id) REFERENCES public.adm_users(usr_id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: adm_members_other_interests adm_fk_mi_usr; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_members_other_interests
    ADD CONSTRAINT adm_fk_mi_usr FOREIGN KEY (moi_usr_id) REFERENCES public.adm_users(usr_id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: membership_sales adm_fk_ms_usr1_id; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.membership_sales
    ADD CONSTRAINT adm_fk_ms_usr1_id FOREIGN KEY (ms_usr1_id) REFERENCES public.adm_users(usr_id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: membership_sales adm_fk_ms_usr2_id; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.membership_sales
    ADD CONSTRAINT adm_fk_ms_usr2_id FOREIGN KEY (ms_usr2_id) REFERENCES public.adm_users(usr_id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: adm_messages_attachments adm_fk_msa_msg_id; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_messages_attachments
    ADD CONSTRAINT adm_fk_msa_msg_id FOREIGN KEY (msa_msg_id) REFERENCES public.adm_messages(msg_id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: adm_messages_content adm_fk_msc_msg_id; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_messages_content
    ADD CONSTRAINT adm_fk_msc_msg_id FOREIGN KEY (msc_msg_id) REFERENCES public.adm_messages(msg_id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: adm_messages_content adm_fk_msc_usr_id; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_messages_content
    ADD CONSTRAINT adm_fk_msc_usr_id FOREIGN KEY (msc_usr_id) REFERENCES public.adm_users(usr_id) ON UPDATE RESTRICT ON DELETE SET NULL;


--
-- Name: adm_messages adm_fk_msg_usr_sender; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_messages
    ADD CONSTRAINT adm_fk_msg_usr_sender FOREIGN KEY (msg_usr_id_sender) REFERENCES public.adm_users(usr_id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: adm_messages_recipients adm_fk_msr_msg_id; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_messages_recipients
    ADD CONSTRAINT adm_fk_msr_msg_id FOREIGN KEY (msr_msg_id) REFERENCES public.adm_messages(msg_id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: adm_messages_recipients adm_fk_msr_rol_id; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_messages_recipients
    ADD CONSTRAINT adm_fk_msr_rol_id FOREIGN KEY (msr_rol_id) REFERENCES public.adm_roles(rol_id) ON UPDATE RESTRICT ON DELETE SET NULL;


--
-- Name: adm_messages_recipients adm_fk_msr_usr_id; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_messages_recipients
    ADD CONSTRAINT adm_fk_msr_usr_id FOREIGN KEY (msr_usr_id) REFERENCES public.adm_users(usr_id) ON UPDATE RESTRICT ON DELETE SET NULL;


--
-- Name: adm_organizations adm_fk_org_org_parent; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_organizations
    ADD CONSTRAINT adm_fk_org_org_parent FOREIGN KEY (org_org_id_parent) REFERENCES public.adm_organizations(org_id) ON UPDATE RESTRICT ON DELETE SET NULL;


--
-- Name: adm_photos adm_fk_pho_org; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_photos
    ADD CONSTRAINT adm_fk_pho_org FOREIGN KEY (pho_org_id) REFERENCES public.adm_organizations(org_id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: adm_photos adm_fk_pho_pho_parent; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_photos
    ADD CONSTRAINT adm_fk_pho_pho_parent FOREIGN KEY (pho_pho_id_parent) REFERENCES public.adm_photos(pho_id) ON UPDATE RESTRICT ON DELETE SET NULL;


--
-- Name: adm_photos adm_fk_pho_usr_change; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_photos
    ADD CONSTRAINT adm_fk_pho_usr_change FOREIGN KEY (pho_usr_id_change) REFERENCES public.adm_users(usr_id) ON UPDATE RESTRICT ON DELETE SET NULL;


--
-- Name: adm_photos adm_fk_pho_usr_create; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_photos
    ADD CONSTRAINT adm_fk_pho_usr_create FOREIGN KEY (pho_usr_id_create) REFERENCES public.adm_users(usr_id) ON UPDATE RESTRICT ON DELETE SET NULL;


--
-- Name: adm_preferences adm_fk_prf_org; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_preferences
    ADD CONSTRAINT adm_fk_prf_org FOREIGN KEY (prf_org_id) REFERENCES public.adm_organizations(org_id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: adm_registrations adm_fk_reg_org; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_registrations
    ADD CONSTRAINT adm_fk_reg_org FOREIGN KEY (reg_org_id) REFERENCES public.adm_organizations(org_id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: adm_registrations adm_fk_reg_usr; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_registrations
    ADD CONSTRAINT adm_fk_reg_usr FOREIGN KEY (reg_usr_id) REFERENCES public.adm_users(usr_id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: adm_role_dependencies adm_fk_rld_rol_child; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_role_dependencies
    ADD CONSTRAINT adm_fk_rld_rol_child FOREIGN KEY (rld_rol_id_child) REFERENCES public.adm_roles(rol_id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: adm_role_dependencies adm_fk_rld_rol_parent; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_role_dependencies
    ADD CONSTRAINT adm_fk_rld_rol_parent FOREIGN KEY (rld_rol_id_parent) REFERENCES public.adm_roles(rol_id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: adm_role_dependencies adm_fk_rld_usr; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_role_dependencies
    ADD CONSTRAINT adm_fk_rld_usr FOREIGN KEY (rld_usr_id) REFERENCES public.adm_users(usr_id) ON UPDATE RESTRICT ON DELETE SET NULL;


--
-- Name: adm_roles adm_fk_rol_cat; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_roles
    ADD CONSTRAINT adm_fk_rol_cat FOREIGN KEY (rol_cat_id) REFERENCES public.adm_categories(cat_id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: adm_roles adm_fk_rol_lst_id; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_roles
    ADD CONSTRAINT adm_fk_rol_lst_id FOREIGN KEY (rol_lst_id) REFERENCES public.adm_lists(lst_id) ON UPDATE SET NULL ON DELETE SET NULL;


--
-- Name: adm_roles adm_fk_rol_usr_change; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_roles
    ADD CONSTRAINT adm_fk_rol_usr_change FOREIGN KEY (rol_usr_id_change) REFERENCES public.adm_users(usr_id) ON UPDATE RESTRICT ON DELETE SET NULL;


--
-- Name: adm_roles adm_fk_rol_usr_create; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_roles
    ADD CONSTRAINT adm_fk_rol_usr_create FOREIGN KEY (rol_usr_id_create) REFERENCES public.adm_users(usr_id) ON UPDATE RESTRICT ON DELETE SET NULL;


--
-- Name: adm_rooms adm_fk_room_usr_change; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_rooms
    ADD CONSTRAINT adm_fk_room_usr_change FOREIGN KEY (room_usr_id_change) REFERENCES public.adm_users(usr_id) ON UPDATE RESTRICT ON DELETE SET NULL;


--
-- Name: adm_rooms adm_fk_room_usr_create; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_rooms
    ADD CONSTRAINT adm_fk_room_usr_create FOREIGN KEY (room_usr_id_create) REFERENCES public.adm_users(usr_id) ON UPDATE RESTRICT ON DELETE SET NULL;


--
-- Name: adm_roles_rights adm_fk_ror_ror_parent; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_roles_rights
    ADD CONSTRAINT adm_fk_ror_ror_parent FOREIGN KEY (ror_ror_id_parent) REFERENCES public.adm_roles_rights(ror_id) ON UPDATE RESTRICT ON DELETE SET NULL;


--
-- Name: adm_roles_rights_data adm_fk_rrd_rol; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_roles_rights_data
    ADD CONSTRAINT adm_fk_rrd_rol FOREIGN KEY (rrd_rol_id) REFERENCES public.adm_roles(rol_id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: adm_roles_rights_data adm_fk_rrd_ror; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_roles_rights_data
    ADD CONSTRAINT adm_fk_rrd_ror FOREIGN KEY (rrd_ror_id) REFERENCES public.adm_roles_rights(ror_id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: adm_roles_rights_data adm_fk_rrd_usr_create; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_roles_rights_data
    ADD CONSTRAINT adm_fk_rrd_usr_create FOREIGN KEY (rrd_usr_id_create) REFERENCES public.adm_users(usr_id) ON UPDATE RESTRICT ON DELETE SET NULL;


--
-- Name: adm_sessions adm_fk_ses_org; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_sessions
    ADD CONSTRAINT adm_fk_ses_org FOREIGN KEY (ses_org_id) REFERENCES public.adm_organizations(org_id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: adm_sessions adm_fk_ses_usr; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_sessions
    ADD CONSTRAINT adm_fk_ses_usr FOREIGN KEY (ses_usr_id) REFERENCES public.adm_users(usr_id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: adm_statistics_columns adm_fk_stc_stt; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_statistics_columns
    ADD CONSTRAINT adm_fk_stc_stt FOREIGN KEY (stc_stt_id) REFERENCES public.adm_statistics_tables(stt_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: adm_statistics_rows adm_fk_str_stt; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_statistics_rows
    ADD CONSTRAINT adm_fk_str_stt FOREIGN KEY (str_stt_id) REFERENCES public.adm_statistics_tables(stt_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: adm_statistics_tables adm_fk_stt_sta; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_statistics_tables
    ADD CONSTRAINT adm_fk_stt_sta FOREIGN KEY (stt_sta_id) REFERENCES public.adm_statistics(sta_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: adm_texts adm_fk_txt_org; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_texts
    ADD CONSTRAINT adm_fk_txt_org FOREIGN KEY (txt_org_id) REFERENCES public.adm_organizations(org_id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: adm_user_relations adm_fk_ure_urt; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_user_relations
    ADD CONSTRAINT adm_fk_ure_urt FOREIGN KEY (ure_urt_id) REFERENCES public.adm_user_relation_types(urt_id) ON UPDATE RESTRICT ON DELETE CASCADE;


--
-- Name: adm_user_relations adm_fk_ure_usr1; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_user_relations
    ADD CONSTRAINT adm_fk_ure_usr1 FOREIGN KEY (ure_usr_id1) REFERENCES public.adm_users(usr_id) ON UPDATE RESTRICT ON DELETE CASCADE;


--
-- Name: adm_user_relations adm_fk_ure_usr2; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_user_relations
    ADD CONSTRAINT adm_fk_ure_usr2 FOREIGN KEY (ure_usr_id2) REFERENCES public.adm_users(usr_id) ON UPDATE RESTRICT ON DELETE CASCADE;


--
-- Name: adm_user_relations adm_fk_ure_usr_change; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_user_relations
    ADD CONSTRAINT adm_fk_ure_usr_change FOREIGN KEY (ure_usr_id_change) REFERENCES public.adm_users(usr_id) ON UPDATE RESTRICT ON DELETE SET NULL;


--
-- Name: adm_user_relations adm_fk_ure_usr_create; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_user_relations
    ADD CONSTRAINT adm_fk_ure_usr_create FOREIGN KEY (ure_usr_id_create) REFERENCES public.adm_users(usr_id) ON UPDATE RESTRICT ON DELETE SET NULL;


--
-- Name: adm_user_relation_types adm_fk_urt_id_inverse; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_user_relation_types
    ADD CONSTRAINT adm_fk_urt_id_inverse FOREIGN KEY (urt_id_inverse) REFERENCES public.adm_user_relation_types(urt_id) ON UPDATE RESTRICT ON DELETE CASCADE;


--
-- Name: adm_user_relation_types adm_fk_urt_usr_change; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_user_relation_types
    ADD CONSTRAINT adm_fk_urt_usr_change FOREIGN KEY (urt_usr_id_change) REFERENCES public.adm_users(usr_id) ON UPDATE RESTRICT ON DELETE SET NULL;


--
-- Name: adm_user_relation_types adm_fk_urt_usr_create; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_user_relation_types
    ADD CONSTRAINT adm_fk_urt_usr_create FOREIGN KEY (urt_usr_id_create) REFERENCES public.adm_users(usr_id) ON UPDATE RESTRICT ON DELETE SET NULL;


--
-- Name: adm_user_data adm_fk_usd_usf; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_user_data
    ADD CONSTRAINT adm_fk_usd_usf FOREIGN KEY (usd_usf_id) REFERENCES public.adm_user_fields(usf_id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: adm_user_data adm_fk_usd_usr; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_user_data
    ADD CONSTRAINT adm_fk_usd_usr FOREIGN KEY (usd_usr_id) REFERENCES public.adm_users(usr_id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: adm_user_log adm_fk_user_log_1; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_user_log
    ADD CONSTRAINT adm_fk_user_log_1 FOREIGN KEY (usl_usr_id) REFERENCES public.adm_users(usr_id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: adm_user_log adm_fk_user_log_2; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_user_log
    ADD CONSTRAINT adm_fk_user_log_2 FOREIGN KEY (usl_usr_id_create) REFERENCES public.adm_users(usr_id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: adm_user_log adm_fk_user_log_3; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_user_log
    ADD CONSTRAINT adm_fk_user_log_3 FOREIGN KEY (usl_usf_id) REFERENCES public.adm_user_fields(usf_id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: adm_user_fields adm_fk_usf_cat; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_user_fields
    ADD CONSTRAINT adm_fk_usf_cat FOREIGN KEY (usf_cat_id) REFERENCES public.adm_categories(cat_id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: adm_user_fields adm_fk_usf_usr_change; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_user_fields
    ADD CONSTRAINT adm_fk_usf_usr_change FOREIGN KEY (usf_usr_id_change) REFERENCES public.adm_users(usr_id) ON UPDATE RESTRICT ON DELETE SET NULL;


--
-- Name: adm_user_fields adm_fk_usf_usr_create; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_user_fields
    ADD CONSTRAINT adm_fk_usf_usr_create FOREIGN KEY (usf_usr_id_create) REFERENCES public.adm_users(usr_id) ON UPDATE RESTRICT ON DELETE SET NULL;


--
-- Name: adm_users adm_fk_usr_usr_change; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_users
    ADD CONSTRAINT adm_fk_usr_usr_change FOREIGN KEY (usr_usr_id_change) REFERENCES public.adm_users(usr_id) ON UPDATE RESTRICT ON DELETE SET NULL;


--
-- Name: adm_users adm_fk_usr_usr_create; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_users
    ADD CONSTRAINT adm_fk_usr_usr_create FOREIGN KEY (usr_usr_id_create) REFERENCES public.adm_users(usr_id) ON UPDATE RESTRICT ON DELETE SET NULL;


--
-- PostgreSQL database dump complete
--

