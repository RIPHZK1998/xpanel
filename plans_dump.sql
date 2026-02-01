--
-- PostgreSQL database dump
--



-- Dumped from database version 15.15 (Homebrew)
-- Dumped by pg_dump version 15.15 (Homebrew)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Data for Name: subscription_plans; Type: TABLE DATA; Schema: public; Owner: postgres
--

INSERT INTO public.subscription_plans (id, name, display_name, duration, price, data_limit_gb, max_devices, status, description, features, created_at, updated_at, deleted_at) VALUES (2, 'weekly_basic', 'Weekly Basic', 'weekly', 5.00, 50, 3, 'active', 'Perfect for short-term needs', '50GB weekly data, 3 devices, Priority support', '2026-01-20 06:05:23.139178-08', '2026-01-20 06:05:23.139178-08', NULL);
INSERT INTO public.subscription_plans (id, name, display_name, duration, price, data_limit_gb, max_devices, status, description, features, created_at, updated_at, deleted_at) VALUES (4, 'quarterly_pro', 'Quarterly Pro', 'quarterly', 40.00, 600, 10, 'active', 'Best value for power users', '600GB quarterly data, 10 devices, Premium support, All nodes, Priority routing', '2026-01-20 06:05:23.140074-08', '2026-01-20 06:05:23.140074-08', NULL);
INSERT INTO public.subscription_plans (id, name, display_name, duration, price, data_limit_gb, max_devices, status, description, features, created_at, updated_at, deleted_at) VALUES (5, 'annual_premium', 'Annual Premium', 'annual', 150.00, 0, 20, 'active', 'Ultimate plan with unlimited data', 'Unlimited data, 20 devices, 24/7 Premium support, All nodes, Priority routing, Dedicated IP option', '2026-01-20 06:05:23.140487-08', '2026-01-20 06:05:23.140487-08', NULL);
INSERT INTO public.subscription_plans (id, name, display_name, duration, price, data_limit_gb, max_devices, status, description, features, created_at, updated_at, deleted_at) VALUES (3, 'monthly_standard', 'Monthly Standard', 'monthly', 15.00, 200, 5, 'active', 'Most popular plan for regular users', '200GB monthly data, 5 devices, Priority support, All nodes', '2026-01-20 06:05:23.139637-08', '2026-01-27 07:40:43.452879-08', NULL);
INSERT INTO public.subscription_plans (id, name, display_name, duration, price, data_limit_gb, max_devices, status, description, features, created_at, updated_at, deleted_at) VALUES (1, 'free', 'Free Plan', 'monthly', 0.00, 5, 2, 'active', 'Basic free plan with limited data', '5GB monthly data, 2 devices, Basic support', '2026-01-20 06:05:23.138223-08', '2026-01-29 21:12:40.994326-08', NULL);


--
-- Name: subscription_plans_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.subscription_plans_id_seq', 5, true);


--
-- PostgreSQL database dump complete
--



