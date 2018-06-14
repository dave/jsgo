package frizz

/*
func (h *Handler) Types(ctx context.Context, info messages.Types, req *http.Request, send func(message services.Message), receive chan services.Message) error {

	s := session.New(nil, assets.Assets, assets.Archives, h.Fileserver, config.ValidExtensions)

	gitreq := h.Cache.NewRequest(true)
	if err := gitreq.InitialiseFromHints(ctx, info.Path); err != nil {
		return err
	}
	g := get.New(s, send, gitreq)

	source, err := getSource(ctx, g, s, info.Path, send)
	if err != nil {
		return err
	}

	if err := s.SetSource(source); err != nil {
		return err
	}

	// set insecure = true in local mode or it will fail if git repo has git protocol
	insecure := config.LOCAL

	// Start the download process - just like the "go get" command.
	if err := g.Get(ctx, info.Path, false, insecure, false); err != nil {
		return err
	}

	if err := gitreq.Close(ctx); err != nil {
		return err
	}

	// Send a message to the client that downloading step has finished.
	send(messages.Downloading{Done: true})

	// Parse for types
	fset := token.NewFileSet()
	bctx := s.BuildContext(false, "")
	parsed := map[string][]*ast.File{}
	for path, files := range source {
		for name, contents := range files {
			if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
				continue
			}
			match, err := bctx.MatchFile(filepath.Join(bctx.GOPATH, "src", path), name)
			if err != nil {
				return err
			}
			if !match {
				continue
			}
			f, err := parser.ParseFile(fset, name, []byte(contents), 0)
			if err != nil {
				return err
			}
			parsed[path] = append(parsed[path], f)
		}
	}
	packages := map[string]*types.Package{}
	tc := types.Config{
		Importer: srcimporter.New(bctx, fset, packages),
	}
	ti := &types.Info{
		Types: map[ast.Expr]types.TypeAndValue{},
		Defs:  map[*ast.Ident]types.Object{},
	}
	p, err := tc.Check(info.Path, fset, parsed[info.Path], ti)
	if err != nil {
		return err
	}

	var globals []gotypes.Named
	for _, v := range ti.Defs {
		if v == nil {
			continue
		}
		if v.Parent() != p.Scope() {
			continue
		}
		if !v.Exported() {
			continue
		}
		tn, ok := v.(*types.TypeName)
		if !ok {
			continue
		}
		n, ok := tn.Type().(*types.Named)
		if !ok {
			return fmt.Errorf("expected *types.Named, got %T", v)
		}
		t := convert.Type(n, &[]types.Type{})
		if t == nil {
			continue
		}
		globals = append(globals, t.(gotypes.Named))
	}
	sort.Slice(globals, func(i, j int) bool { return globals[i].Obj.Name < globals[j].Obj.Name })

	send(messages.TypesComplete{Types: globals})

	return nil
}

func getSource(ctx context.Context, g *get.Getter, s *session.Session, path string, send func(message services.Message)) (map[string]map[string]string, error) {

	root := filepath.Join("goroot", "src", path)
	if _, err := assets.Assets.Stat(root); err == nil {
		// Look in the goroot for standard lib packages
		source, err := getSourceFiles(assets.Assets, path, root)
		if err != nil {
			return nil, err
		}
		send(messages.GetComplete{Source: source})
		return source, nil
	}

	// Send a message to the client that downloading step has started.
	send(messages.Downloading{Starting: true})

	// set insecure = true in local mode or it will fail if git repo has git protocol
	insecure := config.LOCAL

	// Start the download process - just like the "go get" command.
	// Don't need to give git hints here because only one package will be downloaded
	if err := g.Get(ctx, path, false, insecure, true); err != nil {
		return nil, err
	}

	source, err := getSourceFiles(s.GoPath(), path, filepath.Join("gopath", "src", path))
	if err != nil {
		return nil, err
	}

	// Send a message to the client that downloading step has finished.
	send(messages.Downloading{Done: true})
	send(messages.GetComplete{Source: source})

	return source, nil
}

func getSourceFiles(fs billy.Filesystem, path, dir string) (map[string]map[string]string, error) {
	source := map[string]map[string]string{}
	fis, err := fs.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, fi := range fis {
		if !isValidFile(fi.Name()) {
			continue
		}
		if strings.HasSuffix(fi.Name(), "_test.go") {
			continue
		}
		f, err := fs.Open(filepath.Join(dir, fi.Name()))
		if err != nil {
			return nil, err
		}
		b, err := ioutil.ReadAll(f)
		if err != nil {
			f.Close()
			return nil, err
		}
		f.Close()
		if source[path] == nil {
			source[path] = map[string]string{}
		}
		source[path][fi.Name()] = string(b)
	}
	return source, nil
}

func isValidFile(name string) bool {
	for _, ext := range config.ValidExtensions {
		if strings.HasSuffix(name, ext) {
			return true
		}
	}
	return false
}
*/
