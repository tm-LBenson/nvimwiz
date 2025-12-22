local M = {}

local function expand(path)
	if type(path) ~= "string" then
		return ""
	end
	return vim.fn.expand(path)
end

local function list_projects(dir)
	local out = {}
	local paths = vim.fn.globpath(dir, "*", false, true)
	for _, p in ipairs(paths) do
		if vim.fn.isdirectory(p) == 1 then
			local name = vim.fn.fnamemodify(p, ":t")
			table.insert(out, { name = name, path = p })
		end
	end
	table.sort(out, function(a, b) return a.name:lower() < b.name:lower() end)
	return out
end

local function open_project(path)
	if path == nil or path == "" then
		return
	end
	vim.fn.chdir(path)
	vim.cmd("enew")
	local cfg = require("nvimwiz.util").config()
	local explorer = (cfg.choices or {}).explorer
	if explorer == "nvimtree" then
		local ok, api = pcall(require, "nvim-tree.api")
		if ok then
			api.tree.open()
		end
	elseif explorer == "netrw" then
		vim.cmd("Explore")
	end
	local ok_t, tb = pcall(require, "telescope.builtin")
	if ok_t then
		tb.find_files()
	end
end

local function parse_line(line)
	local name = string.match(line or "", "^%s*%d+%.%s+(.*)$")
	if name and name ~= "" then
		return vim.trim(name)
	end
	return nil
end

local function render(buf, dir)
	local projects = list_projects(dir)
	local lines = {}
	table.insert(lines, "Neovim Projects")
	table.insert(lines, "===============")
	table.insert(lines, "")
	table.insert(lines, "Projects dir: " .. dir)
	table.insert(lines, "")
	table.insert(lines, "[p] pick project  [n] new project  [e] explore dir  [r] refresh  [q] quit")
	table.insert(lines, "")
	table.insert(lines, "Projects:")
	for i, prj in ipairs(projects) do
		table.insert(lines, string.format(" %d. %s", i, prj.name))
	end
	table.insert(lines, "")
	table.insert(lines, "Move to a project line and press <Enter> to open.")
	vim.api.nvim_buf_set_lines(buf, 0, -1, false, lines)
	vim.api.nvim_buf_set_option(buf, "modifiable", false)
end

local function show()
	local cfg = require("nvimwiz.util").config()
	local dir = expand(cfg.projects_dir or "~/projects")

	local buf = vim.api.nvim_create_buf(false, true)
	vim.api.nvim_set_current_buf(buf)
	vim.api.nvim_buf_set_option(buf, "buftype", "nofile")
	vim.api.nvim_buf_set_option(buf, "bufhidden", "wipe")
	vim.api.nvim_buf_set_option(buf, "swapfile", false)
	vim.api.nvim_buf_set_option(buf, "modifiable", true)
	render(buf, dir)

	local function refresh()
		vim.api.nvim_buf_set_option(buf, "modifiable", true)
		render(buf, dir)
	end

	local function pick()
		local projects = list_projects(dir)
		local items = {}
		for _, prj in ipairs(projects) do
			table.insert(items, prj)
		end
		vim.ui.select(items, {
			prompt = "Pick project",
			format_item = function(item) return item.name end,
		}, function(item)
			if item then
				open_project(item.path)
			end
		end)
	end

	local function new_project()
		vim.ui.input({ prompt = "New project name" }, function(input)
			local name = vim.trim(input or "")
			if name == "" then
				return
			end
			local path = dir .. "/" .. name
			if vim.fn.isdirectory(path) == 1 then
				open_project(path)
				return
			end
			vim.fn.mkdir(path, "p")
			open_project(path)
		end)
	end

	local function explore()
		vim.cmd("Explore " .. dir)
	end

	local function quit()
		vim.cmd("bd!")
	end

	vim.keymap.set("n", "r", refresh, { buffer = buf, nowait = true })
	vim.keymap.set("n", "p", pick, { buffer = buf, nowait = true })
	vim.keymap.set("n", "n", new_project, { buffer = buf, nowait = true })
	vim.keymap.set("n", "e", explore, { buffer = buf, nowait = true })
	vim.keymap.set("n", "q", quit, { buffer = buf, nowait = true })
	vim.keymap.set("n", "<CR>", function()
		local line = vim.api.nvim_get_current_line()
		local name = parse_line(line)
		if not name then
			return
		end
		open_project(dir .. "/" .. name)
	end, { buffer = buf, nowait = true })
end

function M.setup()
	vim.api.nvim_create_user_command("NvimwizProjects", function() show() end, {})
	vim.api.nvim_create_autocmd("VimEnter", {
		callback = function()
			if vim.fn.argc() ~= 0 then
				return
			end
			local ft = vim.bo.filetype
			if ft ~= "" then
				return
			end
			local line1 = vim.api.nvim_get_current_line()
			if vim.api.nvim_buf_line_count(0) == 1 and (line1 == "" or line1 == nil) then
				show()
			end
		end,
	})
end

return M
