DROP TABLE IF EXISTS verification_applications;
DROP TABLE IF EXISTS verification_badges;
DROP TABLE IF EXISTS invoices;
DROP TABLE IF EXISTS payment_transactions;
DROP TABLE IF EXISTS payment_orders;
DROP TABLE IF EXISTS subscriptions;
DROP TABLE IF EXISTS subscription_plans;
ALTER TABLE posts ALTER COLUMN content TYPE VARCHAR(280);
