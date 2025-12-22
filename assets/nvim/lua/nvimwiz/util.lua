local M = {}

local cfg_cache = nil

function M.config()
	if cfg_cache ~= nil then
		return cfg_cache
	end
	local ok, cfg = pcall(require, "nvimwiz.generated.config")
	if ok and type(cfg) == "table" then
		cfg_cache = cfg
	else
		cfg_cache = {}
	end
	return cfg_cache
end

function M.join(tbls)
	local out = {}
	for _, t in ipairs(tbls or {}) do
		if type(t) == "table" then
			for _, v in ipairs(t) do
				table.insert(out, v)
			end
		end
	end
	return out
end

return M
