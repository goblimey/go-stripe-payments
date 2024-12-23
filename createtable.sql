

--
-- Name: membership_sales; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.membership_sales
(
    ms_id integer NOT NULL,
    ms_payment_service CHARACTER VARYING(36) NOT NULL,
    ms_payment_status CHARACTER VARYING(20) NOT NULL,
    ms_payment_id CHARACTER VARYING(200),
    ms_membership_year integer NOT NULL,
    ms_usr1_id integer NOT NULL,
    ms_usr1_fee REAL NOT NULL,
    ms_usr1_friend boolean NOT NULL,
    ms_usr1_friend_fee REAL,
    ms_usr2_id integer,
    -- null if no associate
    ms_usr2_fee REAL NOT NULL,
    -- 0.0 if no associate
    ms_usr2_friend boolean NOT NULL,
    -- false if no associate
    ms_usr2_friend_fee REAL NOT NULL,
    -- 0.0 if no associate
    ms_donation REAL NOT NULL,
    ms_donation_museum REAL NOT NULL,
    timestamp_create timestamp
    without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
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


    ALTER SEQUENCE public.membership_sales_ms_id_seq
    OWNER TO postgres;

    --
    -- Name: membership_sales_ms_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
    --

    ALTER SEQUENCE public.membership_sales_ms_id_seq
    OWNED BY public.membership_sales.ms_id;


    -- 
    -- Name: membership_sales ms_id; Type: DEFAULT; Schema: public; Owner: postgres
    --

    ALTER TABLE ONLY public.membership_sales
    ALTER COLUMN ms_id
    SET
    DEFAULT nextval
    ('public.membership_sales_ms_id_seq'::regclass);

    --
    -- Name: membership_sales membership_sales_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
    --

    ALTER TABLE ONLY public.membership_sales
    ADD CONSTRAINT membership_sales_pkey PRIMARY KEY
    (ms_id);

    ALTER TABLE ONLY public.membership_sales
    ADD CONSTRAINT adm_fk_ms_usr1_id FOREIGN KEY
    (ms_usr1_id) REFERENCES public.adm_users
    (usr_id) ON
    UPDATE RESTRICT ON
    DELETE RESTRICT;

    ALTER TABLE ONLY public.membership_sales
    ADD CONSTRAINT adm_fk_ms_usr2_id FOREIGN KEY
    (ms_usr2_id) REFERENCES public.adm_users
    (usr_id) ON
    UPDATE RESTRICT ON
    DELETE RESTRICT;