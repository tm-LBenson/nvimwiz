local M = {}

function M.spec()
  return {
    {
      "ThePrimeagen/harpoon",
      dependencies = { "nvim-lua/plenary.nvim" },
      config = function()
        local ok, harpoon = pcall(require, "harpoon")
        if not ok then
          return
        end

        -- Harpoon v1 uses harpoon.mark + harpoon.ui modules.
        harpoon.setup({})

        local mark_ok, mark = pcall(require, "harpoon.mark")
        local ui_ok, ui = pcall(require, "harpoon.ui")
        if not (mark_ok and ui_ok) then
          return
        end

        vim.keymap.set("n", "<leader>ha", mark.add_file, { desc = "Harpoon: add file" })
        vim.keymap.set("n", "<leader>hm", ui.toggle_quick_menu, { desc = "Harpoon: menu" })
        vim.keymap.set("n", "<leader>h1", function()
          ui.nav_file(1)
        end, { desc = "Harpoon: file 1" })
        vim.keymap.set("n", "<leader>h2", function()
          ui.nav_file(2)
        end, { desc = "Harpoon: file 2" })
        vim.keymap.set("n", "<leader>h3", function()
          ui.nav_file(3)
        end, { desc = "Harpoon: file 3" })
        vim.keymap.set("n", "<leader>h4", function()
          ui.nav_file(4)
        end, { desc = "Harpoon: file 4" })
      end,
    },
  }
end

return M
