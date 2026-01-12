
CREATE TABLE IF NOT EXISTS public.adm_countries (
    ct_id INTEGER PRIMARY KEY NOT NULL,
    ct_code CHARACTER VARYING(3) UNIQUE NOT NULL,
    ct_name CHARACTER VARYING(100) UNIQUE NOT NULL
);


ALTER TABLE public.adm_countries OWNER TO postgres;

--
-- Name: adm_countries_ct_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.adm_countries_ct_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.adm_countries_ct_id_seq OWNER TO postgres;

--
-- Name: adm_countries_ct_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.adm_countries_ct_id_seq OWNED BY public.adm_countries.ct_id;

ALTER TABLE ONLY public.adm_countries
ALTER COLUMN ct_id
SET
DEFAULT nextval
('public.adm_countries_ct_id_seq'::regclass);
