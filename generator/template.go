package generator

const CppTemplate = `// Generated by github.com/IronsDu/protoc-gen-gayrpc
// Coding by github.com/liuhan907
// DO NOT EDIT!!!

#ifndef {{$.MacroName}}
#define {{$.MacroName}}

#include <string_view>
#include <string>
#include <unordered_map>
#include <memory>
#include <cstdint>
#include <future>
#include <chrono>

#include <google/protobuf/util/json_util.h>

#include <gayrpc/core/gayrpc_meta.pb.h>
#include "{{$.ModelFileName}}"

#include <gayrpc/core/GayRpcType.h>
#include <gayrpc/core/GayRpcError.h>
#include <gayrpc/core/GayRpcTypeHandler.h>
#include <gayrpc/core/GayRpcClient.h>
#include <gayrpc/core/GayRpcService.h>
#include <gayrpc/core/GayRpcReply.h>
#include <folly/futures/Future.h>

{{range $i, $packageName := $.PackageNames}}namespace {{$packageName}} {
{{end}}
    using namespace gayrpc::core;
    using namespace google::protobuf::util;
    
    enum class {{$.Namespace}}_ServiceID:uint32_t
    {
        {{range $serviceIndex, $service := $.Services}}{{$service.Name}},
        {{end}}
    };

    {{range $serviceIndex, $service := $.Services}}
    enum class {{$service.MethodsEnumName}}:uint64_t
    {
        {{range $i, $method := $service.Methods}}{{$method.EnumName}} = {{$method.MethodID}},
        {{end}}
    };

    class {{$service.Name}}Client : public BaseClient
    {
    public:
        using Ptr = std::shared_ptr<{{$service.Name}}Client>;
        using WeakPtr = std::weak_ptr<{{$service.Name}}Client>;

        {{range $i, $method := $service.Methods}}using {{$method.Name}}Handle = std::function<void(const {{$.ContainerNamespace}}{{$method.OutputType}}&, const std::optional<gayrpc::core::RpcError>&)>;
        {{end}}

    public:
        {{range $i, $method := $service.Methods}}void {{$method.MethodName}}(const {{$.ContainerNamespace}}{{$method.InputType}}& request,
            const {{$method.Name}}Handle& handle = nullptr)
        {
            call<{{$.ContainerNamespace}}{{$method.OutputType}}>(request, 
                static_cast<uint32_t>({{$.Namespace}}_ServiceID::{{$service.Name}}), 
                static_cast<uint64_t>({{$service.MethodsEnumName}}::{{$method.EnumName}}), 
                handle);
        }
        {{end}}
        {{range $i, $method := $service.Methods}}void {{$method.MethodName}}(const {{$.ContainerNamespace}}{{$method.InputType}}& request,
            const {{$method.Name}}Handle& handle,
            std::chrono::seconds timeout, 
            BaseClient::TimeoutCallback&& timeoutCallback)
        {
            call<{{$.ContainerNamespace}}{{$method.OutputType}}>(request, 
                static_cast<uint32_t>({{$.Namespace}}_ServiceID::{{$service.Name}}), 
                static_cast<uint64_t>({{$service.MethodsEnumName}}::{{$method.EnumName}}), 
                handle,
                timeout,
                std::move(timeoutCallback));
        }

        {{end}}
        {{range $i, $method := $service.Methods}}folly::Future<std::pair<{{$.ContainerNamespace}}{{$method.OutputType}}, std::optional<gayrpc::core::RpcError>>> Sync{{$method.MethodName}}(
            const {{$.ContainerNamespace}}{{$method.InputType}}& request,
            std::chrono::seconds timeout)
        {
            auto promise = std::make_shared<folly::Promise<std::pair<{{$.ContainerNamespace}}{{$method.OutputType}}, std::optional<gayrpc::core::RpcError>>>>();

            {{$method.MethodName}}(request, 
                [promise](const {{$.ContainerNamespace}}{{$method.OutputType}}& response,
                    const std::optional<gayrpc::core::RpcError>& error) mutable {
                    promise->setValue(std::make_pair(response, error));
                },
                timeout,
                [promise]() mutable {
                    {{$.ContainerNamespace}}{{$method.OutputType}} response;
                    gayrpc::core::RpcError error;
                    error.setTimeout();
                    promise->setValue(std::make_pair(response, std::optional<gayrpc::core::RpcError>(error)));
                });

            return promise->getFuture();
        }

        {{end}}

        void uninstall()
        {
            getTypeHandleManager()->removeTypeHandle(RpcMeta::RESPONSE, static_cast<uint32_t>({{$.Namespace}}_ServiceID::{{$service.Name}}));
        }

    public:
        static Ptr Create(const RpcTypeHandleManager::Ptr& rpcHandlerManager,
                          const UnaryServerInterceptor& inboundInterceptor,
                          const UnaryServerInterceptor& outboundInterceptor)
        {
            class make_shared_enabler : public {{$service.Name}}Client
            {
            public:
                make_shared_enabler(const RpcTypeHandleManager::Ptr& rpcHandlerManager,
                    const UnaryServerInterceptor& inboundInterceptor,
                    const UnaryServerInterceptor& outboundInterceptor)
                    : 
                    {{$service.Name}}Client(rpcHandlerManager, inboundInterceptor, outboundInterceptor) {}
            };

            auto client = std::make_shared<make_shared_enabler>(rpcHandlerManager, inboundInterceptor, outboundInterceptor);
            client->installResponseStub(rpcHandlerManager, static_cast<uint32_t>({{$.Namespace}}_ServiceID::{{$service.Name}}));

            return client;
        }

        static  std::string GetServiceTypeName()
        {
            return "{{$.ContainerNamespace}}{{$service.Name}}";
        }

    private:
        using BaseClient::BaseClient;
    };

    class {{$service.Name}}Service : public BaseService
    {
    public:
        using Ptr = std::shared_ptr<{{$service.Name}}Service>;
        using WeakPtr = std::weak_ptr<{{$service.Name}}Service>;

        {{range $i, $method := $service.Methods}}using {{$method.Name}}Reply = TemplateReply<{{$.ContainerNamespace}}{{$method.OutputType}}>;
        {{end}}

        using BaseService::BaseService;

        ~{{$service.Name}}Service() override = default;

        void uninstall() final
        {
            getServiceContext().getTypeHandleManager()->removeTypeHandle(RpcMeta::REQUEST, static_cast<uint32_t>({{$.Namespace}}_ServiceID::{{$service.Name}}));
        }

        void install() final
        {
            auto sharedThis = std::static_pointer_cast<{{$service.Name}}Service>(shared_from_this());
            {{$service.Name}}Service::Install(sharedThis);
        }

        static bool Install(const {{$service.Name}}Service::Ptr& service);

        static  std::string GetServiceTypeName()
        {
            return "{{$.ContainerNamespace}}{{$service.Name}}";
        }
    private:
        {{range $i, $method := $service.Methods}}virtual void {{$method.MethodName}}(const {{$.ContainerNamespace}}{{$method.InputType}}& request, 
            const {{$.ContainerNamespace}}{{$service.Name}}Service::{{$method.Name}}Reply::Ptr& replyObj,
            InterceptorContextType&&) = 0;
        {{end}}

    private:

        {{range $i, $method := $service.Methods}}static auto {{$method.MethodName}}_stub(RpcMeta&& meta,
            const std::string_view& data,
            const {{$service.Name}}Service::Ptr& service,
            const UnaryServerInterceptor& inboundInterceptor,
            const UnaryServerInterceptor& outboundInterceptor,
            InterceptorContextType&& context)
        {
            {{$.ContainerNamespace}}{{$method.InputType}} request;
            return parseRequestWrapper(request, std::move(meta), data, inboundInterceptor, [service,
                outboundInterceptor = outboundInterceptor,
                &request](RpcMeta&& meta, const google::protobuf::Message& message, InterceptorContextType&& context) mutable {
                auto replyObject = std::make_shared<{{$method.Name}}Reply>(std::move(meta), std::move(outboundInterceptor));
                service->{{$method.MethodName}}(request, replyObject, std::move(context));
                return gayrpc::core::MakeReadyFuture(std::optional<std::string>(std::nullopt));
            }, std::move(context));
        }

        {{end}}
    };

    inline bool {{$service.Name}}Service::Install(const {{$service.Name}}Service::Ptr& service)
    {
        auto rpcTypeHandleManager = service->getServiceContext().getTypeHandleManager();
        auto inboundInterceptor = service->getServiceContext().getInInterceptor();
        auto outboundInterceptor = service->getServiceContext().getOutInterceptor();

        using {{$service.Name}}ServiceRequestHandler = std::function<InterceptorReturnType(RpcMeta&&,
            const std::string_view& data,
            const {{$service.Name}}Service::Ptr&,
            const UnaryServerInterceptor&,
            const UnaryServerInterceptor&,
            InterceptorContextType&& context)>;

        using {{$service.Name}}ServiceHandlerMapById = std::unordered_map<uint64_t, {{$service.Name}}ServiceRequestHandler>;
        using {{$service.Name}}ServiceHandlerMapByStr = std::unordered_map<std::string, {{$service.Name}}ServiceRequestHandler>;

        {{$service.Name}}ServiceHandlerMapById serviceHandlerMapById = {
            {{range $i, $method := $service.Methods}}{static_cast<uint64_t>({{$service.MethodsEnumName}}::{{$method.MethodName}}), {{$service.Name}}Service::{{$method.MethodName}}_stub},
            {{end}}
        };
        {{$service.Name}}ServiceHandlerMapByStr serviceHandlerMapByStr = {
            {{range $i, $method := $service.Methods}}{"{{range $i, $packageName := $.PackageNames}}{{$packageName}}.{{end}}{{$service.Name}}.{{$method.MethodName}}", {{$service.Name}}Service::{{$method.MethodName}}_stub},
            {{end}}
        };

        auto requestStub = [service,
            serviceHandlerMapById,
            serviceHandlerMapByStr,
            inboundInterceptor,
            outboundInterceptor](RpcMeta&& meta, const std::string_view& data, InterceptorContextType&& context) {

            if (meta.type() != RpcMeta::REQUEST)
            {
                throw std::runtime_error("meta type not request, It is:" + std::to_string(meta.type()));
            }
            
            {{$service.Name}}ServiceRequestHandler handler = nullptr;
            try
            {
                if (!meta.request_info().strmethod().empty())
                {
                    auto it = serviceHandlerMapByStr.find(meta.request_info().strmethod());
                    if (it == serviceHandlerMapByStr.end())
                    {
                        throw std::runtime_error("not found handle, method:" + meta.request_info().strmethod());
                    }
                    handler = (*it).second;
                }
                else
                {
                    auto it = serviceHandlerMapById.find(meta.request_info().intmethod());
                    if (it == serviceHandlerMapById.end())
                    {
                        throw std::runtime_error(std::string("not found handle, method:") + std::to_string(meta.request_info().intmethod()));
                    }
                    handler = (*it).second;
                }
            }
            catch (const std::exception& e)
            {
                BaseReply::ReplyError(outboundInterceptor, meta.service_id(), meta.request_info().sequence_id(), 0, e.what(), InterceptorContextType{});
                return;
            }

            auto future = handler(std::move(meta),
                    data,
                    service,
                    inboundInterceptor,
                    outboundInterceptor,
                    std::move(context));

            if (future.isReady())
            {
                if (future.hasValue())
                {
                    if (auto err = future.value(); err)
                    {
                        BaseReply::ReplyError(outboundInterceptor, meta.service_id(), meta.request_info().sequence_id(), 0, err.value(), InterceptorContextType{});
                    }
                }
                else if(future.hasException())
                {
                    BaseReply::ReplyError(outboundInterceptor, meta.service_id(), meta.request_info().sequence_id(), 0, future.result().exception().get_exception()->what(), InterceptorContextType{});
                    return;
                }
                else
                {
                    throw std::runtime_error("future is ready, but not have any value and exception");
                }
            }
            else
            {
                std::move(future).thenValue([serviceId = meta.service_id(), seqId = meta.request_info().sequence_id(), outboundInterceptor](std::optional<std::string> err) mutable {
                    if (err)
                    {
                        BaseReply::ReplyError(outboundInterceptor, serviceId, seqId, 0, err.value(), InterceptorContextType{});
                    }
                });
            }
        };

        if(!rpcTypeHandleManager->registerTypeHandle(RpcMeta::REQUEST, requestStub, static_cast<uint32_t>({{$.Namespace}}_ServiceID::{{$service.Name}})))
        {
            throw std::runtime_error(std::string("register service:")+ {{$service.Name}}Service::GetServiceTypeName()+" type handler failed");
        }
        return true;
    }
    {{end}}
{{range $i, $packageName := $.PackageNames}}}
{{end}}
#endif

`
