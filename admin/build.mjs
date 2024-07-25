import * as esbuild from "esbuild";

const ADMIN_BUILD_IS_DEV = process.env.ADMIN_BUILD_IS_DEV === "1";

const define = {
  global: "window",
  ...Object.fromEntries(
    Object.entries(process.env)
      .filter(([k, _v]) => k.startsWith("ADMIN_"))
      .map(([k, v]) => [`process.env.${k}`, JSON.stringify(v)]),
  ),
};

const context = await esbuild.context({
  entryPoints: ["./src"],
  outfile: "./public/index.js",
  minify: !ADMIN_BUILD_IS_DEV,
  bundle: true,
  sourcemap: true,
  target: ["chrome58", "firefox57", "safari11", "edge18"],
  define,
});

if (ADMIN_BUILD_IS_DEV) {
  console.log("watching");
  await context.watch();
} else {
  await context.rebuild();
  await context.dispose();
}
