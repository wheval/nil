import { z } from "zod";

export const SignAuthSchema = z.object({
  id: z.number(),
  address: z.string(),
  transaction: z.string(),
  expiredAt: z.number(),
  used: z.boolean(),
});
