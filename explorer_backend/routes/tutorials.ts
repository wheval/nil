import { z } from "zod";
import { router, publicProcedure } from "../trpc";
import { getTutorial, getAllTutorials } from "../services/tutorials_db";

export const tutorialRouter = router({
  get: publicProcedure
    .input(z.number())
    .output(
      z.object({
        id: z.string(),
        text: z.string(),
        contracts: z.string(),
      }),
    )
    .query(async (opts) => {
      const tutorial = await getTutorial(opts.input);
      if (tutorial === null) {
        throw new Error("Tutorial not found");
      }
      return {
        text: tutorial.text,
        contracts: tutorial.contracts,
      };
    }),
  getAll: publicProcedure
    .output(
      z.array(
        z.object({
          id: z.string(),
          text: z.string(),
          contracts: z.string(),
          stage: z.number(),
        }),
      ),
    )
    .query(async () => {
      const tutorials = await getAllTutorials();
      return tutorials;
    }),
});
