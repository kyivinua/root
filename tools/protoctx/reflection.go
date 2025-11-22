package protoctx

import (
	"fmt"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// FindMessageDescriptor finds a message descriptor by fully qualified name.
//
// Example: FindMessageDescriptor("company.user.v1.UserProfile")
func (c *Context) FindMessageDescriptor(fqn string) (protoreflect.MessageDescriptor, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	fullName := protoreflect.FullName(fqn)
	desc, err := c.files.FindDescriptorByName(fullName)
	if err != nil {
		return nil, fmt.Errorf("find descriptor %s: %w", fqn, err)
	}

	md, ok := desc.(protoreflect.MessageDescriptor)
	if !ok {
		return nil, fmt.Errorf("%s is not a message (type: %T)", fqn, desc)
	}

	return md, nil
}

// FindServiceDescriptor finds a service descriptor by fully qualified name.
//
// Example: FindServiceDescriptor("company.user.v1.UserService")
func (c *Context) FindServiceDescriptor(fqn string) (protoreflect.ServiceDescriptor, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	fullName := protoreflect.FullName(fqn)
	desc, err := c.files.FindDescriptorByName(fullName)
	if err != nil {
		return nil, fmt.Errorf("find descriptor %s: %w", fqn, err)
	}

	sd, ok := desc.(protoreflect.ServiceDescriptor)
	if !ok {
		return nil, fmt.Errorf("%s is not a service (type: %T)", fqn, desc)
	}

	return sd, nil
}

// FindEnumDescriptor finds an enum descriptor by fully qualified name.
//
// Example: FindEnumDescriptor("company.user.v1.UserStatus")
func (c *Context) FindEnumDescriptor(fqn string) (protoreflect.EnumDescriptor, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	fullName := protoreflect.FullName(fqn)
	desc, err := c.files.FindDescriptorByName(fullName)
	if err != nil {
		return nil, fmt.Errorf("find descriptor %s: %w", fqn, err)
	}

	ed, ok := desc.(protoreflect.EnumDescriptor)
	if !ok {
		return nil, fmt.Errorf("%s is not an enum (type: %T)", fqn, desc)
	}

	return ed, nil
}

// ListServices returns all service descriptors in the context.
func (c *Context) ListServices() []protoreflect.ServiceDescriptor {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var result []protoreflect.ServiceDescriptor

	c.files.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		for i := 0; i < fd.Services().Len(); i++ {
			result = append(result, fd.Services().Get(i))
		}
		return true
	})

	return result
}

// ListMessages returns all message descriptors in the context (top-level only).
//
// Note: This does not include nested messages. Use DescriptorWalker if you need
// to traverse the full message hierarchy.
func (c *Context) ListMessages() []protoreflect.MessageDescriptor {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var result []protoreflect.MessageDescriptor

	c.files.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		for i := 0; i < fd.Messages().Len(); i++ {
			result = append(result, fd.Messages().Get(i))
		}
		return true
	})

	return result
}

// ListEnums returns all enum descriptors in the context (top-level only).
func (c *Context) ListEnums() []protoreflect.EnumDescriptor {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var result []protoreflect.EnumDescriptor

	c.files.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		for i := 0; i < fd.Enums().Len(); i++ {
			result = append(result, fd.Enums().Get(i))
		}
		return true
	})

	return result
}

// ListServicesByPackage returns all services in a specific package.
func (c *Context) ListServicesByPackage(pkg string) []protoreflect.ServiceDescriptor {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var result []protoreflect.ServiceDescriptor

	c.files.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		if string(fd.Package()) != pkg {
			return true
		}

		for i := 0; i < fd.Services().Len(); i++ {
			result = append(result, fd.Services().Get(i))
		}
		return true
	})

	return result
}

// ListMessagesByPackage returns all messages in a specific package (top-level only).
func (c *Context) ListMessagesByPackage(pkg string) []protoreflect.MessageDescriptor {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var result []protoreflect.MessageDescriptor

	c.files.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		if string(fd.Package()) != pkg {
			return true
		}

		for i := 0; i < fd.Messages().Len(); i++ {
			result = append(result, fd.Messages().Get(i))
		}
		return true
	})

	return result
}

// ListFiles returns all file descriptors in the context.
func (c *Context) ListFiles() []protoreflect.FileDescriptor {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var result []protoreflect.FileDescriptor

	c.files.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		result = append(result, fd)
		return true
	})

	return result
}
