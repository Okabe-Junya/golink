import React from "react"
import ReactDOM from "react-dom/client"
import "./index.css"
import App from "./App"

console.log("Main script executing")

const rootElement = document.getElementById("root")
if (!rootElement) {
  console.error("No root element found")
} else {
  console.log("Root element found, creating React root")
  const root = ReactDOM.createRoot(rootElement)
  root.render(
    <React.StrictMode>
      <App />
    </React.StrictMode>,
  )
}

fetch("/api/links")
  .then((response) => response.json())
  .then((data) => console.log("Fetched links:", data))
  .catch((error) => console.error("Error fetching links:", error))
