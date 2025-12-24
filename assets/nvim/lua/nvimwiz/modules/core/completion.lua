local M = {}

function M.spec()
  return {
    {
      "hrsh7th/nvim-cmp",
      event = "InsertEnter",
      dependencies = {
        "hrsh7th/cmp-nvim-lsp",
        "hrsh7th/cmp-buffer",
        "hrsh7th/cmp-path",
        "saadparwaiz1/cmp_luasnip",
        {
          "L3MON4D3/LuaSnip",
          dependencies = { "rafamadriz/friendly-snippets" },
        },
      },
      config = function()
        vim.opt.completeopt = { "menu", "menuone", "noselect" }
        local cmp = require("cmp")
        local luasnip = require("luasnip")

        pcall(function()
          require("luasnip.loaders.from_vscode").lazy_load()
        end)

        cmp.setup({
          snippet = {
            expand = function(args)
              luasnip.lsp_expand(args.body)
            end,
          },
          mapping = cmp.mapping.preset.insert({
            ["<C-Space>"] = cmp.mapping.complete(),
            ["<CR>"] = cmp.mapping.confirm({ select = true }),
            ["<Tab>"] = cmp.mapping(function(fallback)
              if cmp.visible() then
                cmp.select_next_item()
                return
              end
              if luasnip.expand_or_jumpable() then
                luasnip.expand_or_jump()
                return
              end
              fallback()
            end, { "i", "s" }),
            ["<S-Tab>"] = cmp.mapping(function(fallback)
              if cmp.visible() then
                cmp.select_prev_item()
                return
              end
              if luasnip.jumpable(-1) then
                luasnip.jump(-1)
                return
              end
              fallback()
            end, { "i", "s" }),
          }),
          sources = cmp.config.sources({
            { name = "nvim_lsp" },
            { name = "luasnip" },
            { name = "path" },
          }, {
            { name = "buffer" },
          }),
          window = {
            completion = cmp.config.window.bordered(),
            documentation = cmp.config.window.bordered(),
          },
        })

        local ok, cmp_autopairs = pcall(require, "nvim-autopairs.completion.cmp")
        if ok then
          cmp.event:on("confirm_done", cmp_autopairs.on_confirm_done())
        end
      end,
    },
  }
end

return M
