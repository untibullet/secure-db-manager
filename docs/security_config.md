# Настройка сетевой безопасности

```
==========================================================================
НАСТРОЙКА СЕТЕВОЙ БЕЗОПАСНОСТИ В pg_hba.conf:
==========================================================================

# TYPE  DATABASE        USER                    ADDRESS                 METHOD

# Локальное подключение для администратора
local   all             db_admin                                        scram-sha-256

# Удаленные SSL-подключения с корпоративной сети
hostssl all             db_test_lead            192.168.1.0/24          scram-sha-256
hostssl all             db_automation_engineer  192.168.1.0/24          scram-sha-256
hostssl all             db_tester               192.168.1.0/24          scram-sha-256
hostssl all             db_developer            192.168.1.0/24          scram-sha-256

# CI/CD система с выделенного сервера
hostssl all             db_cicd_system          10.0.10.100/32          scram-sha-256

# Гостевой доступ только через VPN
hostssl all             db_guest                172.16.0.0/16           scram-sha-256

# Запрет всех остальных подключений
host    all             all                     0.0.0.0/0               reject

==========================================================================
НАСТРОЙКА postgresql.conf:
==========================================================================

# Сетевые настройки
listen_addresses = '*'
port = 5432
max_connections = 100

# SSL/TLS обязателен
ssl = on
ssl_cert_file = '/etc/postgresql/ssl/server.crt'
ssl_key_file = '/etc/postgresql/ssl/server.key'
ssl_ca_file = '/etc/postgresql/ssl/ca.crt'
ssl_ciphers = 'HIGH:MEDIUM:+3DES:!aNULL'
ssl_prefer_server_ciphers = on
ssl_min_protocol_version = 'TLSv1.2'

# Аутентификация
password_encryption = scram-sha-256

# Таймауты безопасности
statement_timeout = 300000                    # 5 минут
idle_in_transaction_session_timeout = 600000  # 10 минут
tcp_keepalives_idle = 60
tcp_keepalives_interval = 10
tcp_keepalives_count = 3

# Логирование для аудита
logging_collector = on
log_directory = '/var/log/postgresql'
log_filename = 'postgresql-%Y-%m-%d.log'
log_rotation_age = 1d
log_rotation_size = 100MB

log_connections = on
log_disconnections = on
log_duration = on
log_hostname = on
log_line_prefix = '%t [%p]: [%l-1] user=%u,db=%d,app=%a,client=%h '
log_statement = 'mod'  # Логировать INSERT, UPDATE, DELETE
log_min_duration_statement = 1000  # Логировать запросы > 1 сек

# Защита от перегрузки
shared_buffers = 256MB
work_mem = 8MB
maintenance_work_mem = 128MB
effective_cache_size = 1GB
max_wal_size = 2GB

# Блокировка опасных функций для не-администраторов
default_transaction_read_only = off
```