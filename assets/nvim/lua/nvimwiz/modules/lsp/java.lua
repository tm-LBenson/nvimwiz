local M = {}

function M.spec()
	return {
		{
			"mfussenegger/nvim-jdtls",
			ft = { "java" },
			config = function()
				local cfg = require("nvimwiz.util").config()
				local wants = (cfg.lsp or {}).java
				if not wants then
					return
				end
				local jdtls_bin = vim.fn.exepath("jdtls")
				if jdtls_bin == nil or jdtls_bin == "" then
					vim.notify("jdtls not found on PATH. Open :Mason and install jdtls.", vim.log.levels.WARN)
					return
				end
				local util = require("lspconfig.util")
				local lsp = require("nvimwiz.lsp")
				vim.api.nvim_create_autocmd("FileType", {
					pattern = "java",
					callback = function()
						local root = util.root_pattern(".git", "mvnw", "gradlew", "pom.xml", "build.gradle")(vim.fn.getcwd())
						if root == nil or root == "" then
							root = vim.fn.getcwd()
						end
						local project = vim.fn.fnamemodify(root, ":p:h:t")
						local workspace = vim.fn.stdpath("data") .. "/jdtls-workspace/" .. project
						require("jdtls").start_or_attach({
							cmd = { jdtls_bin, "-data", workspace },
							root_dir = root,
							on_attach = lsp.on_attach,
							capabilities = lsp.capabilities(),
						})
					end,
				})
			end,
		},
	}
end

return M
