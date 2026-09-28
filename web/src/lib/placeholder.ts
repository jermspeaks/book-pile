export function coverHue(title: string) {
  let h = 0;
  for (let i = 0; i < title.length; i++) {
    h = (h * 31 + title.charCodeAt(i)) >>> 0;
  }
  return h % 360;
}
