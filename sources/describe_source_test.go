package sources

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/overmindtech/sdp-go"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestMain(m *testing.M) {
	logrus.SetLevel(logrus.TraceLevel)
	os.Exit(m.Run())
}

func TestType(t *testing.T) {
	s := DescribeOnlySource[string, string, struct{}, struct{}]{
		ItemType: "foo",
	}

	if s.Type() != "foo" {
		t.Errorf("expected type to be foo, got %v", s.Type())
	}
}

func TestName(t *testing.T) {
	// Basically just test that it's not empty. It doesn't matter what it is
	s := DescribeOnlySource[string, string, struct{}, struct{}]{
		ItemType: "foo",
	}

	if s.Name() == "" {
		t.Error("blank name")
	}
}

func TestScopes(t *testing.T) {
	s := DescribeOnlySource[string, string, struct{}, struct{}]{
		Region:    "outer-space",
		AccountID: "mars",
	}

	scopes := s.Scopes()

	if len(scopes) != 1 {
		t.Errorf("expected 1 scope, got %v", len(scopes))
	}

	if scopes[0] != "mars.outer-space" {
		t.Errorf("expected scope to be mars.outer-space, got %v", scopes[0])
	}
}

func TestGet(t *testing.T) {
	t.Run("when everything goes well", func(t *testing.T) {
		var inputMapperCalled bool
		var outputMapperCalled bool
		var describeFuncCalled bool

		s := DescribeOnlySource[string, string, struct{}, struct{}]{
			Region:    "eu-west-2",
			AccountID: "foo",
			InputMapperGet: func(scope, query string) (string, error) {
				inputMapperCalled = true
				return "input", nil
			},
			InputMapperList: func(scope string) (string, error) {
				return "input", nil
			},
			OutputMapper: func(_ context.Context, _ struct{}, scope, input, output string) ([]*sdp.Item, error) {
				outputMapperCalled = true
				return []*sdp.Item{
					{},
				}, nil
			},
			DescribeFunc: func(ctx context.Context, client struct{}, input string) (string, error) {
				describeFuncCalled = true
				return "", nil
			},
		}

		item, err := s.Get(context.Background(), "foo.eu-west-2", "bar", false)

		if err != nil {
			t.Error(err)
		}

		if !inputMapperCalled {
			t.Error("input mapper not called")
		}

		if !outputMapperCalled {
			t.Error("output mapper not called")
		}

		if !describeFuncCalled {
			t.Error("describe func not called")
		}

		if item == nil {
			t.Error("nil item")
		}
	})

	t.Run("use get for list: output returns multiple sources", func(t *testing.T) {
		uniqueAttribute := "virtualGatewayId"
		uniqueAttributeValue := "test-id"

		var inputMapperCalled bool
		var outputMapperCalled bool
		var describeFuncCalled bool

		s := DescribeOnlySource[string, string, struct{}, struct{}]{
			Region:    "eu-west-2",
			AccountID: "foo",
			InputMapperGet: func(scope, query string) (string, error) {
				inputMapperCalled = true
				return "input", nil
			},
			InputMapperList: func(scope string) (string, error) {
				return "input", nil
			},
			OutputMapper: func(_ context.Context, _ struct{}, scope, input, output string) ([]*sdp.Item, error) {
				outputMapperCalled = true
				return []*sdp.Item{
					{
						UniqueAttribute: uniqueAttribute,
						Attributes: &sdp.ItemAttributes{
							AttrStruct: &structpb.Struct{
								Fields: map[string]*structpb.Value{
									uniqueAttribute: structpb.NewStringValue(uniqueAttributeValue),
								},
							},
						},
					},
					{
						UniqueAttribute: uniqueAttribute,
						Attributes: &sdp.ItemAttributes{
							AttrStruct: &structpb.Struct{
								Fields: map[string]*structpb.Value{
									uniqueAttribute: structpb.NewStringValue("some-value"),
								},
							},
						},
					},
				}, nil
			},
			DescribeFunc: func(ctx context.Context, client struct{}, input string) (string, error) {
				describeFuncCalled = true
				return "", nil
			},
			UseListForGet: true,
		}

		item, err := s.Get(context.Background(), "foo.eu-west-2", uniqueAttributeValue, false)

		if err != nil {
			t.Error(err)
		}

		if !inputMapperCalled {
			t.Error("input mapper not called")
		}

		if !outputMapperCalled {
			t.Error("output mapper not called")
		}

		if !describeFuncCalled {
			t.Error("describe func not called")
		}

		if item == nil {
			t.Error("nil item")
		}
	})

	t.Run("with too many results", func(t *testing.T) {
		s := DescribeOnlySource[string, string, struct{}, struct{}]{
			Region:    "eu-west-2",
			AccountID: "foo",
			InputMapperGet: func(scope, query string) (string, error) {
				return "input", nil
			},
			InputMapperList: func(scope string) (string, error) {
				return "input", nil
			},
			OutputMapper: func(_ context.Context, _ struct{}, scope, input, output string) ([]*sdp.Item, error) {
				return []*sdp.Item{
					{},
					{},
					{},
				}, nil
			},
			DescribeFunc: func(ctx context.Context, client struct{}, input string) (string, error) {
				return "", nil
			},
		}

		_, err := s.Get(context.Background(), "foo.eu-west-2", "bar", false)

		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("with no results", func(t *testing.T) {
		s := DescribeOnlySource[string, string, struct{}, struct{}]{
			Region:    "eu-west-2",
			AccountID: "foo",
			InputMapperGet: func(scope, query string) (string, error) {
				return "input", nil
			},
			InputMapperList: func(scope string) (string, error) {
				return "input", nil
			},
			OutputMapper: func(_ context.Context, _ struct{}, scope, input, output string) ([]*sdp.Item, error) {
				return []*sdp.Item{}, nil
			},
			DescribeFunc: func(ctx context.Context, client struct{}, input string) (string, error) {
				return "", nil
			},
		}

		_, err := s.Get(context.Background(), "foo.eu-west-2", "bar", false)

		if err == nil {
			t.Error("expected error")
		}
	})
}

func TestSearchARN(t *testing.T) {
	s := DescribeOnlySource[string, string, struct{}, struct{}]{
		Region:    "region",
		AccountID: "account-id",
		InputMapperGet: func(scope, query string) (string, error) {
			return "input", nil
		},
		InputMapperList: func(scope string) (string, error) {
			return "input", nil
		},
		OutputMapper: func(_ context.Context, _ struct{}, scope, input, output string) ([]*sdp.Item, error) {
			return []*sdp.Item{
				{},
			}, nil
		},
		DescribeFunc: func(ctx context.Context, client struct{}, input string) (string, error) {
			return "fancy", nil
		},
	}

	items, err := s.Search(context.Background(), "account-id.region", "arn:partition:service:region:account-id:resource-type:resource-id", false)

	if err != nil {
		t.Error(err)
	}

	if len(items) != 1 {
		t.Errorf("expected 1 item, got %v", len(items))
	}
}

func TestSearchCustom(t *testing.T) {
	s := DescribeOnlySource[string, string, struct{}, struct{}]{
		Region:    "region",
		AccountID: "account-id",
		InputMapperGet: func(scope, query string) (string, error) {
			return "input", nil
		},
		InputMapperList: func(scope string) (string, error) {
			return "input", nil
		},
		OutputMapper: func(_ context.Context, _ struct{}, scope, input, output string) ([]*sdp.Item, error) {
			return []*sdp.Item{
				{
					Type:            "test-item",
					UniqueAttribute: "name",
					Attributes: &sdp.ItemAttributes{
						AttrStruct: &structpb.Struct{
							Fields: map[string]*structpb.Value{
								"name": structpb.NewStringValue(output),
							},
						},
					},
				},
			}, nil
		},
		InputMapperSearch: func(ctx context.Context, client struct{}, scope, query string) (string, error) {
			return "custom", nil
		},
		DescribeFunc: func(ctx context.Context, client struct{}, input string) (string, error) {
			return input, nil
		},
	}

	items, err := s.Search(context.Background(), "account-id.region", "foo", false)

	if err != nil {
		t.Fatal(err)
	}

	if len(items) != 1 {
		t.Errorf("expected 1 item, got %v", len(items))
	}

	if items[0].UniqueAttributeValue() != "custom" {
		t.Errorf("expected item to be 'custom', got %v", items[0].UniqueAttributeValue())
	}
}

func TestNoInputMapper(t *testing.T) {
	s := DescribeOnlySource[string, string, struct{}, struct{}]{
		Region:    "eu-west-2",
		AccountID: "foo",
		OutputMapper: func(_ context.Context, _ struct{}, scope, input, output string) ([]*sdp.Item, error) {
			return []*sdp.Item{
				{},
			}, nil
		},
		DescribeFunc: func(ctx context.Context, client struct{}, input string) (string, error) {
			return "", nil
		},
	}

	t.Run("Get", func(t *testing.T) {
		_, err := s.Get(context.Background(), "foo.eu-west-2", "bar", false)

		if err == nil {
			t.Error("expected error but got nil")
		}
	})

	t.Run("List", func(t *testing.T) {
		_, err := s.List(context.Background(), "foo.eu-west-2", false)

		if err == nil {
			t.Error("expected error but got nil")
		}
	})
}

func TestNoOutputMapper(t *testing.T) {
	s := DescribeOnlySource[string, string, struct{}, struct{}]{
		Region:    "eu-west-2",
		AccountID: "foo",
		InputMapperGet: func(scope, query string) (string, error) {
			return "input", nil
		},
		InputMapperList: func(scope string) (string, error) {
			return "input", nil
		},
		DescribeFunc: func(ctx context.Context, client struct{}, input string) (string, error) {
			return "", nil
		},
	}

	t.Run("Get", func(t *testing.T) {
		_, err := s.Get(context.Background(), "foo.eu-west-2", "bar", false)

		if err == nil {
			t.Error("expected error but got nil")
		}
	})

	t.Run("List", func(t *testing.T) {
		_, err := s.List(context.Background(), "foo.eu-west-2", false)

		if err == nil {
			t.Error("expected error but got nil")
		}
	})
}

func TestNoDescribeFunc(t *testing.T) {
	s := DescribeOnlySource[string, string, struct{}, struct{}]{
		Region:    "eu-west-2",
		AccountID: "foo",
		InputMapperGet: func(scope, query string) (string, error) {
			return "input", nil
		},
		InputMapperList: func(scope string) (string, error) {
			return "input", nil
		},
		OutputMapper: func(_ context.Context, _ struct{}, scope, input, output string) ([]*sdp.Item, error) {
			return []*sdp.Item{
				{},
			}, nil
		},
	}

	t.Run("Get", func(t *testing.T) {
		_, err := s.Get(context.Background(), "foo.eu-west-2", "bar", false)

		if err == nil {
			t.Error("expected error but got nil")
		}
	})

	t.Run("List", func(t *testing.T) {
		_, err := s.List(context.Background(), "foo.eu-west-2", false)

		if err == nil {
			t.Error("expected error but got nil")
		}
	})
}

func TestFailingInputMapper(t *testing.T) {
	s := DescribeOnlySource[string, string, struct{}, struct{}]{
		Region:    "eu-west-2",
		AccountID: "foo",
		InputMapperGet: func(scope, query string) (string, error) {
			return "input", errors.New("foobar")
		},
		InputMapperList: func(scope string) (string, error) {
			return "input", errors.New("foobar")
		},
		OutputMapper: func(_ context.Context, _ struct{}, scope, input, output string) ([]*sdp.Item, error) {
			return []*sdp.Item{
				{},
			}, nil
		},
		DescribeFunc: func(ctx context.Context, client struct{}, input string) (string, error) {
			return "", nil
		},
	}

	fooBar := regexp.MustCompile("foobar")

	t.Run("Get", func(t *testing.T) {
		_, err := s.Get(context.Background(), "foo.eu-west-2", "bar", false)

		if err == nil {
			t.Error("expected error but got nil")
		}

		if !fooBar.MatchString(err.Error()) {
			t.Errorf("expected error string '%v' to contain foobar", err.Error())
		}
	})

	t.Run("List", func(t *testing.T) {
		_, err := s.List(context.Background(), "foo.eu-west-2", false)

		if err == nil {
			t.Error("expected error but got nil")
		}

		if !fooBar.MatchString(err.Error()) {
			t.Errorf("expected error string '%v' to contain foobar", err.Error())
		}
	})
}

func TestFailingOutputMapper(t *testing.T) {
	s := DescribeOnlySource[string, string, struct{}, struct{}]{
		Region:    "eu-west-2",
		AccountID: "foo",
		InputMapperGet: func(scope, query string) (string, error) {
			return "input", nil
		},
		InputMapperList: func(scope string) (string, error) {
			return "input", nil
		},
		OutputMapper: func(_ context.Context, _ struct{}, scope, input, output string) ([]*sdp.Item, error) {
			return nil, errors.New("foobar")
		},
		DescribeFunc: func(ctx context.Context, client struct{}, input string) (string, error) {
			return "", nil
		},
	}

	fooBar := regexp.MustCompile("foobar")

	t.Run("Get", func(t *testing.T) {
		_, err := s.Get(context.Background(), "foo.eu-west-2", "bar", false)

		if err == nil {
			t.Error("expected error but got nil")
		}

		if !fooBar.MatchString(err.Error()) {
			t.Errorf("expected error string '%v' to contain foobar", err.Error())
		}
	})

	t.Run("List", func(t *testing.T) {
		_, err := s.List(context.Background(), "foo.eu-west-2", false)

		if err == nil {
			t.Error("expected error but got nil")
		}

		if !fooBar.MatchString(err.Error()) {
			t.Errorf("expected error string '%v' to contain foobar", err.Error())
		}
	})
}

func TestFailingDescribeFunc(t *testing.T) {
	s := DescribeOnlySource[string, string, struct{}, struct{}]{
		Region:    "eu-west-2",
		AccountID: "foo",
		InputMapperGet: func(scope, query string) (string, error) {
			return "input", nil
		},
		InputMapperList: func(scope string) (string, error) {
			return "input", nil
		},
		OutputMapper: func(_ context.Context, _ struct{}, scope, input, output string) ([]*sdp.Item, error) {
			return []*sdp.Item{
				{},
			}, nil
		},
		DescribeFunc: func(ctx context.Context, client struct{}, input string) (string, error) {
			return "", errors.New("foobar")
		},
	}

	fooBar := regexp.MustCompile("foobar")

	t.Run("Get", func(t *testing.T) {
		_, err := s.Get(context.Background(), "foo.eu-west-2", "bar", false)

		if err == nil {
			t.Error("expected error but got nil")
		}

		if !fooBar.MatchString(err.Error()) {
			t.Errorf("expected error string '%v' to contain foobar", err.Error())
		}
	})

	t.Run("List", func(t *testing.T) {
		_, err := s.List(context.Background(), "foo.eu-west-2", false)

		if err == nil {
			t.Error("expected error but got nil")
		}

		if !fooBar.MatchString(err.Error()) {
			t.Errorf("expected error string '%v' to contain foobar", err.Error())
		}
	})
}

type TestPaginator struct {
	DataFunc func() string

	MaxPages int

	page int
}

func (t *TestPaginator) HasMorePages() bool {
	if t.MaxPages == 0 {
		t.MaxPages = 3
	}
	return t.page < t.MaxPages
}

func (t *TestPaginator) NextPage(context.Context, ...func(struct{})) (string, error) {
	data := t.DataFunc()
	t.page++
	return data, nil
}

func TestPaginated(t *testing.T) {
	s := DescribeOnlySource[string, string, struct{}, struct{}]{
		MaxResultsPerPage: 1,
		Region:            "eu-west-2",
		AccountID:         "foo",
		InputMapperGet: func(scope, query string) (string, error) {
			return "input", nil
		},
		InputMapperList: func(scope string) (string, error) {
			return "input", nil
		},
		OutputMapper: func(_ context.Context, _ struct{}, scope, input, output string) ([]*sdp.Item, error) {
			return []*sdp.Item{
				{},
			}, nil
		},
		PaginatorBuilder: func(client struct{}, params string) Paginator[string, struct{}] {
			return &TestPaginator{DataFunc: func() string {
				return "foo"
			}}
		},
		DescribeFunc: func(ctx context.Context, client struct{}, input string) (string, error) {
			return "", nil
		},
	}

	t.Run("detecting pagination", func(t *testing.T) {
		if !s.Paginated() {
			t.Error("pagination not detected")
		}

		if err := s.Validate(); err != nil {
			t.Error(err)
		}
	})

	t.Run("paginating a List query", func(t *testing.T) {
		items, err := s.List(context.Background(), "foo.eu-west-2", false)

		if err != nil {
			t.Error(err)
		}

		if len(items) != 3 {
			t.Errorf("expected 3 items, got %v", len(items))
		}
	})
}

func TestDescribeOnlySourceCaching(t *testing.T) {
	ctx := context.Background()
	generation := 0
	s := DescribeOnlySource[string, string, struct{}, struct{}]{
		ItemType:          "test-type",
		MaxResultsPerPage: 1,
		Region:            "eu-west-2",
		AccountID:         "foo",
		InputMapperGet: func(scope, query string) (string, error) {
			return "input", nil
		},
		InputMapperList: func(scope string) (string, error) {
			return "input", nil
		},
		OutputMapper: func(_ context.Context, _ struct{}, scope, input, output string) ([]*sdp.Item, error) {
			return []*sdp.Item{
				{
					Scope:           "foo.eu-west-2",
					Type:            "test-type",
					UniqueAttribute: "name",
					Attributes: &sdp.ItemAttributes{
						AttrStruct: &structpb.Struct{
							Fields: map[string]*structpb.Value{
								"name":       structpb.NewStringValue("test-item"),
								"generation": structpb.NewStringValue(output),
							},
						},
					},
				},
			}, nil
		},
		PaginatorBuilder: func(client struct{}, params string) Paginator[string, struct{}] {
			return &TestPaginator{
				DataFunc: func() string {
					generation += 1
					return fmt.Sprintf("%v", generation)
				},
				MaxPages: 1,
			}
		},
		DescribeFunc: func(ctx context.Context, client struct{}, input string) (string, error) {
			generation += 1
			return fmt.Sprintf("%v", generation), nil
		},
	}

	t.Run("get", func(t *testing.T) {
		// get
		first, err := s.Get(ctx, "foo.eu-west-2", "test-item", false)
		if err != nil {
			t.Fatal(err)
		}
		firstGen, err := first.GetAttributes().Get("generation")
		if err != nil {
			t.Fatal(err)
		}

		// get again
		withCache, err := s.Get(ctx, "foo.eu-west-2", "test-item", false)
		if err != nil {
			t.Fatal(err)
		}
		withCacheGen, err := withCache.GetAttributes().Get("generation")
		if err != nil {
			t.Fatal(err)
		}

		if firstGen != withCacheGen {
			t.Errorf("with cache: expected generation %v, got %v", firstGen, withCacheGen)
		}

		// get ignore cache
		withoutCache, err := s.Get(ctx, "foo.eu-west-2", "test-item", true)
		if err != nil {
			t.Fatal(err)
		}
		withoutCacheGen, err := withoutCache.GetAttributes().Get("generation")
		if err != nil {
			t.Fatal(err)
		}
		if withoutCacheGen == firstGen {
			t.Errorf("with cache: expected generation %v, got %v", firstGen, withoutCacheGen)
		}
	})

	t.Run("list", func(t *testing.T) {
		// list
		first, err := s.List(ctx, "foo.eu-west-2", false)
		if err != nil {
			t.Fatal(err)
		}
		firstGen, err := first[0].GetAttributes().Get("generation")
		if err != nil {
			t.Fatal(err)
		}

		// list again
		withCache, err := s.List(ctx, "foo.eu-west-2", false)
		if err != nil {
			t.Fatal(err)
		}
		withCacheGen, err := withCache[0].GetAttributes().Get("generation")
		if err != nil {
			t.Fatal(err)
		}

		if firstGen != withCacheGen {
			t.Errorf("with cache: expected generation %v, got %v", firstGen, withCacheGen)
		}

		// list ignore cache
		withoutCache, err := s.List(ctx, "foo.eu-west-2", true)
		if err != nil {
			t.Fatal(err)
		}
		withoutCacheGen, err := withoutCache[0].GetAttributes().Get("generation")
		if err != nil {
			t.Fatal(err)
		}

		if withoutCacheGen == firstGen {
			t.Errorf("with cache: expected generation %v, got %v", firstGen, withoutCacheGen)
		}
	})

	t.Run("search", func(t *testing.T) {
		// search
		first, err := s.Search(ctx, "foo.eu-west-2", "arn:aws:test-type:eu-west-2:foo:test-item", false)
		if err != nil {
			t.Fatal(err)
		}
		firstGen, err := first[0].GetAttributes().Get("generation")
		if err != nil {
			t.Fatal(err)
		}

		// search again
		withCache, err := s.Search(ctx, "foo.eu-west-2", "arn:aws:test-type:eu-west-2:foo:test-item", false)
		if err != nil {
			t.Fatal(err)
		}
		withCacheGen, err := withCache[0].GetAttributes().Get("generation")
		if err != nil {
			t.Fatal(err)
		}

		if firstGen != withCacheGen {
			t.Errorf("with cache: expected generation %v, got %v", firstGen, withCacheGen)
		}

		// search ignore cache
		withoutCache, err := s.Search(ctx, "foo.eu-west-2", "arn:aws:test-type:eu-west-2:foo:test-item", true)
		if err != nil {
			t.Fatal(err)
		}
		withoutCacheGen, err := withoutCache[0].GetAttributes().Get("generation")
		if err != nil {
			t.Fatal(err)
		}
		if withoutCacheGen == firstGen {
			t.Errorf("with cache: expected generation %v, got %v", firstGen, withoutCacheGen)
		}
	})
}
