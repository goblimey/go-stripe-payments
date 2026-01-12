-- This script sets up the extra tables used to store the members' interests. 
--
-- adm_interests contains a fixed set of possible interests.  It's used to
-- populate a selection list.
--
-- adm_members_interests contains the members' interests, one to many from adm_users
-- to adm_interests, so the possible contents is restricted.
--
-- adm_members_other_interests contains member's interests that are not included in
-- adm_interests.  It's free text.  Administrators should keep an eye on this table 
-- and update adm_interests when lots of users express the same extra interest.   

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


ALTER TABLE public.adm_interests_ntrst_id_seq OWNER TO postgres;

--
-- Name: adm_interests; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.adm_interests (
    ntrst_id integer DEFAULT nextval('public.adm_interests_ntrst_id_seq'::regclass) NOT NULL,
    ntrst_name character varying(50)
);


ALTER TABLE public.adm_interests OWNER TO postgres;


--
-- Data for Name: adm_interests; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.adm_interests (ntrst_id, ntrst_name) FROM stdin;
1	Archaeology
2	Archives
3	Buildings
4	Cartography
5	Creative Arts
6	Cultural Heritage
7	History
8	Military History
9	Natural History
10	Social History
11	People
12	Transport
\.



--
-- Name: adm_interests_ntrst_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.adm_interests_ntrst_id_seq', 13, true);






--
-- Name: adm_members_interests; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.adm_members_interests (
    mi_id integer NOT NULL,
    mi_usr_id integer NOT NULL,
    mi_interest_id integer NOT NULL,
    CONSTRAINT adm_un_usr_interest UNIQUE (mi_usr_id, mi_interest_id)
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


ALTER TABLE public.adm_members_interests_mi_id_seq OWNER TO postgres;


--
-- Name: adm_members_interests_mi_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.adm_members_interests_mi_id_seq OWNED BY public.adm_members_interests.mi_id;


--
-- Name: adm_members_interests_mi_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.adm_members_interests_mi_id_seq', 1, true);


--
-- Name: adm_interests adm_interests_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_interests
    ADD CONSTRAINT adm_interests_pkey PRIMARY KEY (ntrst_id);





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


ALTER TABLE public.adm_members_other_interests_moi_id_seq OWNER TO postgres;


--
-- Name: adm_members_other_interests_moi_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.adm_members_other_interests_moi_id_seq OWNED BY public.adm_members_other_interests.moi_id;


--
-- Name: adm_members_interests mi_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_members_interests ALTER COLUMN mi_id SET DEFAULT nextval('public.adm_members_interests_mi_id_seq'::regclass);


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
-- Name: adm_members_other_interests moi_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_members_other_interests ALTER COLUMN moi_id SET DEFAULT nextval('public.adm_members_other_interests_moi_id_seq'::regclass);


--
-- Name: adm_members_other_interests_moi_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.adm_members_other_interests_moi_id_seq', 1, true);


--
-- Name: adm_members_interests adm_members_interests_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_members_interests
    ADD CONSTRAINT adm_members_interests_pkey PRIMARY KEY (mi_id);

--
-- Name: adm_members_other_interests adm_fk_mi_usr; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.adm_members_other_interests
    ADD CONSTRAINT adm_fk_mi_usr FOREIGN KEY (moi_usr_id) REFERENCES public.adm_users(usr_id) ON UPDATE RESTRICT ON DELETE RESTRICT;



