import { match } from "ts-pattern"
import { z } from "zod"

const _emptyStringToUndefined = z.literal("").transform(() => undefined)

export const asOptional = <T>(schema: z.ZodType<T>) =>
  schema.optional().or(_emptyStringToUndefined)

export const zodErrorMap = (issue: z.core.$ZodIssue) => {
  const message = match(issue)
    .with({ code: "invalid_type" }, (issue) => {
      if (issue.input === undefined || issue.input === null || Number.isNaN(issue.input)) {
        return "This field is required."
      }
    })
    .with({ code: "invalid_format", format: "email" }, () =>
      "Please input a valid email address."
    )
    .with({ code: "invalid_format", format: "url" }, () =>
      "Please input a valid URL."
    )
    .with({ code: "too_small" }, (issue) => {
      if (issue.input === "") return "This field is required."
    })
    .otherwise(() => undefined)

  const result = message || issue.message
  return result.endsWith(".") ? result : `${result}.`
}

export { z }
