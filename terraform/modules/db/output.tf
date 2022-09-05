output "db_instance" {
  value = google_sql_database_instance.main.connection_name
}
