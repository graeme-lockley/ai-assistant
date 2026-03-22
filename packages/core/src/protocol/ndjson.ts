export class NDJSONWriter {
  constructor(private readonly out: { write: (s: string) => void }) {}

  writeLine(obj: Record<string, unknown>): void {
    this.out.write(JSON.stringify(obj) + "\n");
  }
}
