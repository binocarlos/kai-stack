// Generate a @solana/kit-compatible TypeScript client from the Anchor IDL.
//
// Pipeline: anchor build -> program/target/idl/counter.json -> (this script) ->
// clients/ts/src/generated/{accounts,instructions,programs,errors,pdas,...}.
//
// The output is committed to the repo as reference, but is fully regenerable:
//   ./stack codegen
//
// Read the IDL with fs (instead of an `import ... with { type: 'json' }`
// assertion) so this runs the same on any Node 18/20/22.
import { readFileSync, rmSync } from "node:fs";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";

import { createFromRoot } from "codama";
import { rootNodeFromAnchor } from "@codama/nodes-from-anchor";
import { renderVisitor } from "@codama/renderers-js";

const here = dirname(fileURLToPath(import.meta.url));
const idlPath = resolve(here, "../../program/target/idl/counter.json");
const outDir = resolve(here, "src/generated");

const idl = JSON.parse(readFileSync(idlPath, "utf-8"));
const codama = createFromRoot(rootNodeFromAnchor(idl));
// generatedFolder: "." writes the modules flat into outDir (otherwise the
// renderer nests them under outDir/src/generated and scaffolds a package).
// The render visitor is async — await it before touching its output.
await codama.accept(renderVisitor(outDir, { generatedFolder: "." }));

// The renderer also scaffolds a standalone package.json. We consume the client
// as source from the monorepo, where clients/ts/package.json ("type":"module")
// applies. Leaving this stray package.json (which lacks "type":"module") would
// flip the generated files to CommonJS and break their named ESM exports.
rmSync(resolve(outDir, "package.json"), { force: true });

console.log(`Generated @solana/kit client -> ${outDir}`);
