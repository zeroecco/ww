package anchor

import (
	"context"

	api "github.com/wetware/ww/internal/api/cluster"
)

type ContainerAnchorImpl struct {
	path []string
	cap  api.Anchor_Container
}

func (dai ContainerAnchorImpl) Path() []string {
	return dai.path
}

func (dai ContainerAnchorImpl) Set(ctx context.Context, data []byte) error {
	fut, release := dai.cap.Set(ctx, func(a api.Anchor_Container_set_Params) error {
		return a.SetData(data)
	})
	defer release()

	select {
	case <-fut.Done():
	case <-ctx.Done():
		return ctx.Err()
	}

	_, err := fut.Struct()
	return err
}

func (dai ContainerAnchorImpl) Get(ctx context.Context) ([]byte, error) {
	fut, release := dai.cap.Get(ctx, func(a api.Anchor_Container_get_Params) error {
		return nil
	})
	defer release()

	select {
	case <-fut.Done():
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	results, err := fut.Struct()
	if err != nil {
		return nil, err
	}
	buffer, err := results.Data()
	if err != nil {
		return nil, err
	}

	data := make([]byte, len(buffer))
	copy(data, buffer)
	return data, nil
}
