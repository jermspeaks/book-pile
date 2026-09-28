export function ownershipFromFlags(physical: boolean, digital: boolean) {
  if (physical && digital) return "both";
  if (physical) return "physical";
  if (digital) return "digital";
  return "coveting";
}

export function ownershipLabel(ownership: string) {
  switch (ownership) {
    case "physical":
      return "Physical";
    case "digital":
      return "Digital";
    case "both":
      return "Physical + digital";
    default:
      return "Coveting";
  }
}
