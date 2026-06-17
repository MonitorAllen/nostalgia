export const htmlToPlainText = (html: string) => {
  const container = document.createElement('div')
  container.innerHTML = html
  return container.textContent?.trim() || ''
}

export const writeRichClipboard = async (html: string, plainText = htmlToPlainText(html)) => {
  if (html && navigator.clipboard?.write && typeof ClipboardItem !== 'undefined' && window.isSecureContext) {
    await navigator.clipboard.write([
      new ClipboardItem({
        'text/html': new Blob([html], { type: 'text/html' }),
        'text/plain': new Blob([plainText], { type: 'text/plain' })
      })
    ])
    return
  }

  if (navigator.clipboard?.writeText && window.isSecureContext) {
    await navigator.clipboard.writeText(plainText)
    return
  }

  const textarea = document.createElement('textarea')
  textarea.value = plainText
  textarea.setAttribute('readonly', 'true')
  textarea.style.position = 'fixed'
  textarea.style.left = '-9999px'
  textarea.style.top = '0'
  document.body.appendChild(textarea)
  textarea.focus()
  textarea.select()
  const copied = document.execCommand('copy')
  textarea.remove()
  if (!copied) {
    throw new Error('copy command failed')
  }
}
