-- Rollback: Remove tabela de planos e coluna plano_id das lojas

DROP INDEX IF EXISTS idx_lojas_plano_id;

ALTER TABLE lojas
    DROP COLUMN IF EXISTS plano_id;

DROP INDEX IF EXISTS idx_planos_is_active;
DROP INDEX IF EXISTS idx_planos_slug;

DROP TABLE IF EXISTS planos;
