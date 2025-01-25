import { Command } from "@oclif/core";

export default class SmartAccount extends Command {
  static override description = "Interact with the smart account set in the config file";

  async run(): Promise<void> {
    await this.config.runCommand("help", ["smart-account"]);
  }
}
