-- public.ai_models definition

-- Drop table

-- DROP TABLE public.ai_models;

CREATE TABLE public.ai_models (
	id text NOT NULL,
	user_id text DEFAULT 'default'::text NOT NULL,
	"name" text NOT NULL,
	provider text NOT NULL,
	enabled bool DEFAULT false NULL,
	api_key text DEFAULT ''::text NULL,
	custom_api_url text DEFAULT ''::text NULL,
	custom_model_name text DEFAULT ''::text NULL,
	created_at timestamp DEFAULT CURRENT_TIMESTAMP NULL,
	updated_at timestamp DEFAULT CURRENT_TIMESTAMP NULL,
	CONSTRAINT ai_models_pkey PRIMARY KEY (id, user_id)
);


-- public.api_keys definition

-- Drop table

-- DROP TABLE public.api_keys;

CREATE TABLE public.api_keys (
	id text NOT NULL,
	user_id text NOT NULL,
	key_hash text NOT NULL,
	key_prefix text NOT NULL,
	"name" text DEFAULT 'Default Key'::text NULL,
	enabled bool DEFAULT true NULL,
	rate_limit int4 DEFAULT 1000 NULL,
	usage_count int4 DEFAULT 0 NULL,
	last_used_at timestamp NULL,
	created_at timestamp DEFAULT CURRENT_TIMESTAMP NULL,
	expires_at timestamp NULL,
	CONSTRAINT api_keys_key_hash_key UNIQUE (key_hash),
	CONSTRAINT api_keys_pkey PRIMARY KEY (id)
);


-- public.api_usage_logs definition

-- Drop table

-- DROP TABLE public.api_usage_logs;

CREATE TABLE public.api_usage_logs (
	id serial4 NOT NULL,
	api_key_id text NOT NULL,
	user_id text NOT NULL,
	endpoint text NOT NULL,
	"method" text NOT NULL,
	status_code int4 NULL,
	response_time_ms int4 NULL,
	created_at timestamp DEFAULT CURRENT_TIMESTAMP NULL,
	CONSTRAINT api_usage_logs_pkey PRIMARY KEY (id)
);


-- public.beta_codes definition

-- Drop table

-- DROP TABLE public.beta_codes;

CREATE TABLE public.beta_codes (
	code text NOT NULL,
	used bool DEFAULT false NULL,
	used_by text DEFAULT ''::text NULL,
	used_at timestamp NULL,
	created_at timestamp DEFAULT CURRENT_TIMESTAMP NULL,
	CONSTRAINT beta_codes_pkey PRIMARY KEY (code)
);


-- public.decision_logs definition

-- Drop table

-- DROP TABLE public.decision_logs;

CREATE TABLE public.decision_logs (
	id text NOT NULL,
	user_id text NOT NULL,
	trader_id text NOT NULL,
	cycle_number int4 NOT NULL,
	"timestamp" timestamp NOT NULL,
	system_prompt text DEFAULT ''::text NULL,
	input_prompt text DEFAULT ''::text NULL,
	cot_trace text DEFAULT ''::text NULL,
	account_state text NOT NULL,
	positions text DEFAULT ''::text NULL,
	decisions text DEFAULT ''::text NULL,
	candidate_coins text DEFAULT ''::text NULL,
	execution_log text DEFAULT ''::text NULL,
	success bool DEFAULT false NULL,
	error_message text DEFAULT ''::text NULL,
	ai_request_duration_ms int4 DEFAULT 0 NULL,
	created_at timestamp DEFAULT CURRENT_TIMESTAMP NULL,
	prompt_tokens int4 DEFAULT 0 NULL,
	completion_tokens int4 DEFAULT 0 NULL,
	total_tokens int4 DEFAULT 0 NULL,
	CONSTRAINT decision_logs_pkey PRIMARY KEY (id)
);
CREATE INDEX idx_decision_logs_timestamp ON public.decision_logs USING btree ("timestamp" DESC);
CREATE INDEX idx_decision_logs_trader_id_timestamp ON public.decision_logs USING btree (trader_id, "timestamp");
CREATE INDEX idx_decision_logs_user_trader ON public.decision_logs USING btree (user_id, trader_id, "timestamp" DESC);


-- public.equity_history definition

-- Drop table

-- DROP TABLE public.equity_history;

CREATE TABLE public.equity_history (
	id serial4 NOT NULL,
	user_id text NOT NULL,
	trader_id text NOT NULL,
	"timestamp" timestamp NOT NULL,
	total_equity float4 NOT NULL,
	available_balance float4 NOT NULL,
	total_unrealized_profit float4 NOT NULL,
	total_pnl float4 NOT NULL,
	total_pnl_pct float4 NOT NULL,
	position_count int4 DEFAULT 0 NULL,
	margin_used_pct float4 DEFAULT 0 NULL,
	created_at timestamp DEFAULT CURRENT_TIMESTAMP NULL,
	CONSTRAINT equity_history_pkey PRIMARY KEY (id)
);
CREATE INDEX idx_equity_history_timestamp ON public.equity_history USING btree ("timestamp" DESC);
CREATE INDEX idx_equity_history_user_trader ON public.equity_history USING btree (user_id, trader_id, "timestamp" DESC);


-- public.exchanges definition

-- Drop table

-- DROP TABLE public.exchanges;

CREATE TABLE public.exchanges (
	id text NOT NULL,
	user_id text DEFAULT 'default'::text NOT NULL,
	api_key_name text NOT NULL,
	"name" text NOT NULL,
	"type" text NOT NULL,
	enabled bool DEFAULT false NULL,
	api_key text DEFAULT ''::text NULL,
	secret_key text DEFAULT ''::text NULL,
	passphrase text DEFAULT ''::text NULL,
	testnet bool DEFAULT false NULL,
	hyperliquid_wallet_addr text DEFAULT ''::text NULL,
	aster_user text DEFAULT ''::text NULL,
	aster_signer text DEFAULT ''::text NULL,
	aster_private_key text DEFAULT ''::text NULL,
	created_at timestamp DEFAULT CURRENT_TIMESTAMP NULL,
	updated_at timestamp DEFAULT CURRENT_TIMESTAMP NULL,
	CONSTRAINT exchanges_pkey PRIMARY KEY (id, user_id, api_key_name)
);


-- public.system_ai_models definition

-- Drop table

-- DROP TABLE public.system_ai_models;

CREATE TABLE public.system_ai_models (
	id text NOT NULL,
	"name" text NOT NULL,
	provider text NOT NULL,
	description text DEFAULT ''::text NULL,
	default_model_name text DEFAULT ''::text NULL,
	default_api_url text DEFAULT ''::text NULL,
	created_at timestamp DEFAULT CURRENT_TIMESTAMP NULL,
	CONSTRAINT system_ai_models_pkey PRIMARY KEY (id)
);


-- public.system_config definition

-- Drop table

-- DROP TABLE public.system_config;

CREATE TABLE public.system_config (
	"key" text NOT NULL,
	value text NOT NULL,
	updated_at timestamp DEFAULT CURRENT_TIMESTAMP NULL,
	CONSTRAINT system_config_pkey PRIMARY KEY (key)
);


-- public.system_exchanges definition

-- Drop table

-- DROP TABLE public.system_exchanges;

CREATE TABLE public.system_exchanges (
	id text NOT NULL,
	"name" text NOT NULL,
	"type" text NOT NULL,
	description text DEFAULT ''::text NULL,
	supports_testnet bool DEFAULT false NULL,
	created_at timestamp DEFAULT CURRENT_TIMESTAMP NULL,
	CONSTRAINT system_exchanges_pkey PRIMARY KEY (id)
);


-- public.traders definition

-- Drop table

-- DROP TABLE public.traders;

CREATE TABLE public.traders (
	id text NOT NULL,
	user_id text DEFAULT 'default'::text NOT NULL,
	"name" text NOT NULL,
	ai_model_id text NOT NULL,
	exchange_id text NOT NULL,
	exchange_api_key_name text DEFAULT ''::text NULL,
	initial_balance float4 NOT NULL,
	scan_interval_minutes int4 DEFAULT 3 NULL,
	is_running bool DEFAULT false NULL,
	btc_eth_leverage int4 DEFAULT 5 NULL,
	altcoin_leverage int4 DEFAULT 5 NULL,
	trading_symbols text DEFAULT ''::text NULL,
	use_coin_pool bool DEFAULT false NULL,
	use_oi_top bool DEFAULT false NULL,
	custom_prompt text DEFAULT ''::text NULL,
	override_base_prompt bool DEFAULT false NULL,
	is_cross_margin bool DEFAULT true NULL,
	use_default_coins bool DEFAULT true NULL,
	custom_coins text DEFAULT ''::text NULL,
	system_prompt_template text DEFAULT 'default'::text NULL,
	created_at timestamp DEFAULT CURRENT_TIMESTAMP NULL,
	updated_at timestamp DEFAULT CURRENT_TIMESTAMP NULL,
	total_tokens int8 DEFAULT 0 NULL,
	is_deleted text DEFAULT 'n'::text NULL,
	CONSTRAINT traders_pkey PRIMARY KEY (id)
);
CREATE INDEX idx_traders_is_deleted ON public.traders USING btree (is_deleted);
CREATE INDEX idx_traders_user_id_is_deleted ON public.traders USING btree (user_id, is_deleted);


-- public.user_signal_sources definition

-- Drop table

-- DROP TABLE public.user_signal_sources;

CREATE TABLE public.user_signal_sources (
	id serial4 NOT NULL,
	user_id text NOT NULL,
	coin_pool_url text DEFAULT ''::text NULL,
	oi_top_url text DEFAULT ''::text NULL,
	created_at timestamp DEFAULT CURRENT_TIMESTAMP NULL,
	updated_at timestamp DEFAULT CURRENT_TIMESTAMP NULL,
	CONSTRAINT user_signal_sources_pkey PRIMARY KEY (id),
	CONSTRAINT user_signal_sources_user_id_key UNIQUE (user_id)
);


-- public.user_trading_config definition

-- Drop table

-- DROP TABLE public.user_trading_config;

CREATE TABLE public.user_trading_config (
	user_id text NOT NULL,
	btc_eth_leverage int4 NULL,
	altcoin_leverage int4 NULL,
	max_daily_loss float4 NULL,
	max_drawdown float4 NULL,
	stop_trading_minutes int4 NULL,
	use_default_coins bool DEFAULT true NULL,
	custom_coins text DEFAULT ''::text NULL,
	created_at timestamp DEFAULT CURRENT_TIMESTAMP NULL,
	updated_at timestamp DEFAULT CURRENT_TIMESTAMP NULL,
	CONSTRAINT user_trading_config_pkey PRIMARY KEY (user_id)
);


-- public.users definition

-- Drop table

-- DROP TABLE public.users;

CREATE TABLE public.users (
	id text NOT NULL,
	email text NOT NULL,
	password_hash text NOT NULL,
	otp_secret text NULL,
	otp_verified bool DEFAULT false NULL,
	is_admin bool DEFAULT false NULL,
	created_at timestamp DEFAULT CURRENT_TIMESTAMP NULL,
	updated_at timestamp DEFAULT CURRENT_TIMESTAMP NULL,
	total_tokens int8 DEFAULT 0 NULL,
	CONSTRAINT users_email_key UNIQUE (email),
	CONSTRAINT users_pkey PRIMARY KEY (id)
);