export const AVATAR_MAX_UPLOAD_BYTES = 2 * 1024 * 1024
export const AVATAR_MAX_OUTPUT_BYTES = 512 * 1024
export const AVATAR_OUTPUT_SIZE = 512
export const AVATAR_ALLOWED_TYPES = ['image/jpeg', 'image/png', 'image/webp']

export function validateAvatarFile(file: File): string | null {
  if (!AVATAR_ALLOWED_TYPES.includes(file.type)) {
    return 'invalid-type'
  }

  if (file.size > AVATAR_MAX_UPLOAD_BYTES) {
    return 'too-large'
  }

  return null
}

export function loadImageDataUrl(file: Blob): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(String(reader.result || ''))
    reader.onerror = () => reject(new Error('Failed to read file'))
    reader.readAsDataURL(file)
  })
}

export type AvatarCropParams = {
  image: HTMLImageElement
  scale: number
  offsetX: number
  offsetY: number
  previewSize: number
}

function canvasToBlob(canvas: HTMLCanvasElement, quality: number): Promise<Blob> {
  return new Promise((resolve, reject) => {
    canvas.toBlob((blob) => {
      if (!blob) {
        reject(new Error('Failed to export image'))
        return
      }
      resolve(blob)
    }, 'image/jpeg', quality)
  })
}

export async function cropAvatarToDataUrl({
  image,
  scale,
  offsetX,
  offsetY,
  previewSize,
}: AvatarCropParams): Promise<string> {
  const imageRatio = image.naturalWidth / image.naturalHeight
  const baseScaleToPreview =
    imageRatio >= 1 ? previewSize / image.naturalHeight : previewSize / image.naturalWidth
  const totalScale = baseScaleToPreview * scale

  const sourceWidth = previewSize / totalScale
  const sourceHeight = previewSize / totalScale
  const unclampedX = (image.naturalWidth - sourceWidth) / 2 - offsetX / totalScale
  const unclampedY = (image.naturalHeight - sourceHeight) / 2 - offsetY / totalScale
  const sourceX = Math.max(0, Math.min(image.naturalWidth - sourceWidth, unclampedX))
  const sourceY = Math.max(0, Math.min(image.naturalHeight - sourceHeight, unclampedY))

  const canvas = document.createElement('canvas')
  canvas.width = AVATAR_OUTPUT_SIZE
  canvas.height = AVATAR_OUTPUT_SIZE
  const ctx = canvas.getContext('2d')
  if (!ctx) {
    throw new Error('Failed to create canvas')
  }

  ctx.fillStyle = '#ffffff'
  ctx.fillRect(0, 0, canvas.width, canvas.height)
  ctx.drawImage(
    image,
    sourceX,
    sourceY,
    sourceWidth,
    sourceHeight,
    0,
    0,
    canvas.width,
    canvas.height,
  )

  for (const quality of [0.92, 0.84, 0.76, 0.68]) {
    const blob = await canvasToBlob(canvas, quality)
    if (blob.size <= AVATAR_MAX_OUTPUT_BYTES) {
      return await loadImageDataUrl(blob)
    }
  }

  throw new Error('Avatar image is too large after cropping')
}
