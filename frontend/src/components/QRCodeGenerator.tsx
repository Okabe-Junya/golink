import type React from "react"
import { useState } from "react"
import { QRCodeCanvas as QRCode } from "qrcode.react"

interface QRCodeGeneratorProps {
  url: string
  size?: number
  includeMargin?: boolean
  logoUrl?: string
}

const QRCodeGenerator: React.FC<QRCodeGeneratorProps> = ({
  url,
  size = 200,
  includeMargin = true,
  logoUrl,
}) => {
  const [bgColor, setBgColor] = useState<string>("#FFFFFF")
  const [fgColor, setFgColor] = useState<string>("#000000")
  const [qrLevel, setQrLevel] = useState<"L" | "M" | "Q" | "H">("M")
  const [downloadName, setDownloadName] = useState<string>("qrcode")

  // Generate canvas download link
  const downloadQR = () => {
    const canvas = document.getElementById("qr-code") as HTMLCanvasElement
    if (!canvas) return

    // Convert canvas to data URL and create download link
    const pngUrl = canvas
      .toDataURL("image/png")
      .replace("image/png", "image/octet-stream")

    const downloadLink = document.createElement("a")
    downloadLink.href = pngUrl
    downloadLink.download = `${downloadName || "qrcode"}.png`
    document.body.appendChild(downloadLink)
    downloadLink.click()
    document.body.removeChild(downloadLink)
  }

  return (
    <div className="card bg-base-200 shadow-xl p-6">
      <div className="flex flex-col md:flex-row gap-6">
        <div className="flex flex-col items-center justify-center bg-white p-4 rounded-lg">
          <QRCode
            id="qr-code"
            value={url}
            size={size}
            bgColor={bgColor}
            fgColor={fgColor}
            level={qrLevel}
            includeMargin={includeMargin}
            imageSettings={
              logoUrl
                ? {
                    src: logoUrl,
                    x: undefined,
                    y: undefined,
                    height: 24,
                    width: 24,
                    excavate: true,
                  }
                : undefined
            }
          />
          <p className="mt-2 text-center text-sm text-base-content">
            Scan to access: <br />
            <a
              href={url}
              target="_blank"
              rel="noopener noreferrer"
              className="link link-primary break-all"
            >
              {url}
            </a>
          </p>
        </div>

        <div className="form-control w-full max-w-xs">
          <h3 className="text-lg font-bold mb-4">QR Code Options</h3>

          <label htmlFor="bg-color" className="label">
            <span className="label-text">Background Color</span>
          </label>
          <input
            id="bg-color"
            type="color"
            value={bgColor}
            onChange={(e) => setBgColor(e.target.value)}
            className="input input-bordered w-full max-w-xs h-10"
          />

          <label htmlFor="fg-color" className="label">
            <span className="label-text">Foreground Color</span>
          </label>
          <input
            id="fg-color"
            type="color"
            value={fgColor}
            onChange={(e) => setFgColor(e.target.value)}
            className="input input-bordered w-full max-w-xs h-10"
          />

          <label htmlFor="qr-level" className="label">
            <span className="label-text">Error Correction Level</span>
          </label>
          <select
            id="qr-level"
            value={qrLevel}
            onChange={(e) =>
              setQrLevel(e.target.value as "L" | "M" | "Q" | "H")
            }
            className="select select-bordered w-full max-w-xs"
          >
            <option value="L">Low (7%)</option>
            <option value="M">Medium (15%)</option>
            <option value="Q">Quartile (25%)</option>
            <option value="H">High (30%)</option>
          </select>

          <label htmlFor="download-name" className="label">
            <span className="label-text">Download Filename</span>
          </label>
          <input
            id="download-name"
            type="text"
            value={downloadName}
            onChange={(e) => setDownloadName(e.target.value)}
            placeholder="QR code filename"
            className="input input-bordered w-full max-w-xs"
          />

          <button
            type="button"
            onClick={downloadQR}
            className="btn btn-primary mt-4"
          >
            Download QR Code
          </button>
        </div>
      </div>
    </div>
  )
}

export default QRCodeGenerator
