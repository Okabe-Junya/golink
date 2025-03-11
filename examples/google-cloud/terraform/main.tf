# Enable required Google Cloud APIs
resource "google_project_service" "required_apis" {
  for_each = toset([
    "run.googleapis.com",
    "artifactregistry.googleapis.com",
    "cloudresourcemanager.googleapis.com",
    "iam.googleapis.com",
    "compute.googleapis.com",
    "firestore.googleapis.com"
  ])

  service            = each.key
  disable_on_destroy = false
}

resource "google_service_account" "cloud_run_service_account" {
  account_id   = "golink-service-account"
  display_name = "GoLink Service Account"
  depends_on   = [google_project_service.required_apis]
}

resource "google_cloud_run_v2_service" "frontend" {
  name     = var.frontend_service_name
  location = var.region

  template {
    scaling {
      min_instance_count = var.min_instance_count
      max_instance_count = var.max_instance_count
    }

    containers {
      image = var.frontend_container_image

      resources {
        limits = {
          cpu    = "1"
          memory = "512Mi"
        }
      }

      env {
        name  = "REACT_APP_API_URL"
        value = "https://${google_cloud_run_v2_service.backend.uri}/api"
      }

      env {
        name  = "REACT_APP_DOMAIN"
        value = google_cloud_run_v2_service.frontend.uri
      }
    }

    service_account = google_service_account.cloud_run_service_account.email
  }

  depends_on = [google_project_service.required_apis]
}

# Create a random session key if not provided
resource "random_password" "session_key" {
  count   = var.session_secret_key == "" ? 1 : 0
  length  = 32
  special = false
}

resource "google_cloud_run_v2_service" "backend" {
  name     = var.backend_service_name
  location = var.region

  template {
    scaling {
      min_instance_count = var.min_instance_count
      max_instance_count = var.max_instance_count
    }

    containers {
      image = var.backend_container_image

      resources {
        limits = {
          cpu    = "1"
          memory = "512Mi"
        }
      }

      env {
        name  = "PORT"
        value = "8080"
      }

      env {
        name  = "APP_DOMAIN"
        value = google_cloud_run_v2_service.backend.uri
      }

      env {
        name  = "CORS_ORIGIN"
        value = "https://${google_cloud_run_v2_service.frontend.uri}"
      }

      env {
        name  = "FRONTEND_URL"
        value = "https://${google_cloud_run_v2_service.frontend.uri}"
      }

      # Authentication environment variables
      env {
        name  = "GOOGLE_CLIENT_ID"
        value = var.auth_enabled ? var.google_client_id : ""
      }

      env {
        name  = "GOOGLE_CLIENT_SECRET"
        value = var.auth_enabled ? var.google_client_secret : ""
      }

      env {
        name  = "GOOGLE_ALLOWED_DOMAIN"
        value = var.auth_enabled ? var.google_allowed_domain : ""
      }

      env {
        name  = "SESSION_SECRET_KEY"
        value = var.auth_enabled ? (var.session_secret_key != "" ? var.session_secret_key : random_password.session_key[0].result) : ""
      }

      env {
        name  = "OAUTH_REDIRECT_URL"
        value = var.auth_enabled ? "https://${google_cloud_run_v2_service.backend.uri}/api/auth/callback" : ""
      }
    }

    service_account = google_service_account.cloud_run_service_account.email
  }

  depends_on = [google_project_service.required_apis]
}

resource "google_cloud_run_v2_service_iam_member" "frontend_public" {
  location = google_cloud_run_v2_service.frontend.location
  name     = google_cloud_run_v2_service.frontend.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}

resource "google_cloud_run_v2_service_iam_member" "backend_public" {
  location = google_cloud_run_v2_service.backend.location
  name     = google_cloud_run_v2_service.backend.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}

resource "google_firestore_database" "database" {
  project     = var.project_id
  name        = "golink-db"
  location_id = var.region
  type        = "FIRESTORE_NATIVE"

  depends_on = [google_project_service.required_apis]
}

resource "google_project_iam_binding" "firestore_access" {
  project = var.project_id
  role    = "roles/datastore.user"

  members = [
    "serviceAccount:${google_service_account.cloud_run_service_account.email}"
  ]

  depends_on = [google_project_service.required_apis, google_firestore_database.database]
}
