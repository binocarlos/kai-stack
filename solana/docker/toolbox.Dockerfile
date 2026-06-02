# The "toolbox": the official Anchor image (Rust + Agave/Solana CLI + Anchor)
# with Node 22 on top. The image pins Node 20.18 via nvm, but @solana/kit's deps
# (@noble/hashes) require Node >= 20.19 — so we install Node 22 from NodeSource
# and make it win over the bundled nvm Node. One container then runs
# `anchor build/deploy`, `anchor test` and the Codama codegen.
#
# Anchor 1.0's `anchor test` spawns `surfpool` as its local validator, so we
# copy that (glibc) binary out of the official surfpool image.
FROM surfpool/surfpool:latest AS surfpool

FROM solanafoundation/anchor:v1.0.2

COPY --from=surfpool /bin/surfpool /usr/local/bin/surfpool

RUN apt-get update \
 && apt-get install -y --no-install-recommends curl ca-certificates gnupg \
 && curl -fsSL https://deb.nodesource.com/setup_22.x | bash - \
 && apt-get install -y --no-install-recommends nodejs \
 && apt-get clean \
 && rm -rf /var/lib/apt/lists/*

# NodeSource installs to /usr/bin; put it ahead of the nvm Node 20 on PATH.
# Solana/Anchor/Cargo bins live in their own dirs, so they still resolve.
ENV PATH=/usr/bin:$PATH

RUN npm install -g yarn

WORKDIR /workspace/program
