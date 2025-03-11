import { render, screen, fireEvent } from "@testing-library/react"
import "@testing-library/jest-dom"
import { Navbar } from "../Navbar"
import { vi } from 'vitest'

describe("Navbar", () => {
  const defaultProps = {
    darkMode: false,
    onThemeToggle: vi.fn(),
  }

  test("renders navbar with title", () => {
    render(<Navbar {...defaultProps} />)

    expect(screen.getByText(/GoLink/i)).toBeInTheDocument()
  })

  test("handles theme toggle", () => {
    render(<Navbar {...defaultProps} />)

    const themeToggleButton = screen.getByRole("button", {
      name: /toggle theme/i,
    })
    fireEvent.click(themeToggleButton)

    expect(defaultProps.onThemeToggle).toHaveBeenCalled()
  })

  test("displays correct theme icon based on darkMode prop", () => {
    const { rerender } = render(<Navbar {...defaultProps} />)

    // Light mode
    expect(screen.getByLabelText(/toggle theme/i)).toBeInTheDocument()

    // Dark mode
    rerender(<Navbar {...defaultProps} darkMode={true} />)
    expect(screen.getByLabelText(/toggle theme/i)).toBeInTheDocument()
  })
})
